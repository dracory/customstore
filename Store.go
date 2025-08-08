package customstore

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	"github.com/dracory/database"
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/sb"
	"github.com/samber/lo"
)

// ============================================================================
// == CLASS
// ============================================================================

// Store defines a session store
type storeImplementation struct {
	tableName          string
	db                 *sql.DB
	dbDriverName       string
	automigrateEnabled bool
	debugEnabled       bool
	logger             *slog.Logger
}

// ============================================================================
// == CONSTRUCTOR
// ============================================================================

// NewStoreOptions define the options for creating a new session store
type NewStoreOptions struct {
	TableName          string
	DB                 *sql.DB
	DbDriverName       string
	TimeoutSeconds     int64
	AutomigrateEnabled bool
	DebugEnabled       bool
	Logger             *slog.Logger
}

// ============================================================================
// == METHODS
// ============================================================================

// NewStore creates a new session store
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	store := &storeImplementation{
		tableName:          opts.TableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
		logger:             opts.Logger,
	}

	if store.tableName == "" {
		return nil, errors.New("customstore store: tableName is required")
	}

	if store.db == nil {
		return nil, errors.New("session store: DB is required")
	}

	if store.dbDriverName == "" {
		store.dbDriverName = sb.DatabaseDriverName(store.db)
	}

	if store.logger == nil {
		store.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	if store.automigrateEnabled {
		store.AutoMigrate()
	}

	return store, nil
}

// AutoMigrate migrates the tables
func (st *storeImplementation) AutoMigrate() error {
	sql := st.SqlCreateTable()

	if st.debugEnabled {
		log.Println(sql)
	}

	_, err := st.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *storeImplementation) EnableDebug(debugEnabled bool) {
	st.debugEnabled = debugEnabled
}

// RecordCreate creates a record
// func (st *storeImplementation) RecordCreate(record RecordInterface) error {
// 	if record.ID() == "" {
// 		record.SetID(uid.HumanUid())
// 	}

// 	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
// 	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

// 	var sqlStr string
// 	sqlStr, sqlParams, errSQL := goqu.Dialect(st.dbDriverName).
// 		Insert(st.tableName).
// 		Rows(record).
// 		Prepared(true).
// 		ToSQL()

// 	if errSQL != nil {
// 		return errSQL
// 	}

// 	if st.debugEnabled {
// 		log.Println(sqlStr)
// 	}

// 	_, err := st.db.Exec(sqlStr, sqlParams...)

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// RecordFindByID finds a record by ID
// func (st *storeImplementation) RecordFindByID(id string) (RecordInterface, error) {
// 	sqlStr, sqlParams, _ := goqu.Dialect(st.dbDriverName).
// 		From(st.tableName).
// 		Prepared(true).
// 		Where(goqu.C(COLUMN_ID).Eq(id), goqu.C(COLUMN_SOFT_DELETED_AT).IsNull()).
// 		Limit(1).
// 		Select().
// 		ToSQL()

// 	if st.debugEnabled {
// 		log.Println(sqlStr)
// 	}

// 	var record recordImplementation
// 	err := sqlscan.Get(context.Background(), st.db, &record, sqlStr, sqlParams...)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			// Looks like this is now outdated for sqlscan
// 			return nil, nil
// 		}
// 		if sqlscan.NotFound(err) {
// 			return nil, nil
// 		}
// 		log.Println("Failed to execute query: ", err)
// 		return nil, err
// 	}

// 	return &record, nil
// }

// RecordUpdate updates a record
// func (st *storeImplementation) RecordUpdate(record RecordInterface) error {
// 	fields := map[string]interface{}{}
// 	fields[COLUMN_PAYLOAD] = record.Payload()
// 	fields[COLUMN_UPDATED_AT] = time.Now()

// 	var sqlStr string

// 	sqlStr, sqlParams, errSQL := goqu.Dialect(st.dbDriverName).
// 		Update(st.tableName).
// 		Set(fields).
// 		Where(goqu.C("id").Eq(record.ID)).
// 		Prepared(true).
// 		ToSQL()

// 	if errSQL != nil {
// 		return errSQL
// 	}

// 	if st.debugEnabled {
// 		log.Println(sqlStr)
// 	}

// 	_, err := st.db.Exec(sqlStr, sqlParams...)

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// RecordCount counts the number of records that match the query
func (st *storeImplementation) RecordCount(options RecordQueryInterface) (int64, error) {
	if st.db == nil {
		return 0, errors.New("database is not initialized")
	}

	options.SetCountOnly(true)

	q, _, err := options.ToSelectDataset(st.dbDriverName, st.tableName)

	if err != nil {
		return -1, err
	}

	sqlStr, sqlParams, err := q.
		Prepared(true).
		Limit(1).
		Select(goqu.COUNT(goqu.Star()).As("count")).
		ToSQL()

	if err != nil {
		return -1, err
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	mapped, err := database.SelectToMapString(database.Context(context.Background(), st.db), sqlStr, sqlParams...)
	if err != nil {
		return -1, err
	}

	if len(mapped) < 1 {
		return -1, nil
	}

	countStr := mapped[0]["count"]

	count, err := strconv.ParseInt(countStr, 10, 64)

	if err != nil {
		return -1, err
	}

	return count, nil
}

// RecordCreate creates a new record
func (st *storeImplementation) RecordCreate(record RecordInterface) error {
	if st.db == nil {
		return errors.New("database is not initialized")
	}

	if record.ID() == "" {
		return errors.New("record ID is required")
	}

	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := record.Data()

	sqlStr, sqlParams, err := goqu.Dialect(st.dbDriverName).
		Insert(st.tableName).
		Prepared(true).
		Rows(data).
		ToSQL()

	if err != nil {
		return err
	}

	if st.debugEnabled {
		st.logger.Debug("Record create query", "query", sqlStr, "params", sqlParams)
	}

	_, err = database.Execute(database.Context(context.Background(), st.db), sqlStr, sqlParams...)

	if err != nil {
		return err
	}

	record.MarkAsNotDirty()

	return nil
}

// RecordDelete permanently deletes a record
func (st *storeImplementation) RecordDelete(record RecordInterface) error {
	if record == nil {
		return errors.New("record is nil")
	}

	return st.RecordDeleteByID(record.ID())
}

// RecordDeleteByID permanently deletes a record by ID
func (st *storeImplementation) RecordDeleteByID(id string) error {
	if st.db == nil {
		return errors.New("database is not initialized")
	}

	if id == "" {
		return errors.New("record id is empty")
	}

	sqlStr, sqlParams, err := goqu.Dialect(st.dbDriverName).
		Delete(st.tableName).
		Prepared(true).
		Where(goqu.C(COLUMN_ID).Eq(id)).
		ToSQL()

	if err != nil {
		return err
	}

	if st.debugEnabled {
		st.logger.Debug("Incident delete query", "query", sqlStr, "params", sqlParams)
	}

	_, err = database.Execute(database.Context(context.Background(), st.db), sqlStr, sqlParams...)
	if err != nil {
		return err
	}

	return nil
}

// RecordFindByID returns a record by ID
func (st *storeImplementation) RecordFindByID(id string) (record RecordInterface, err error) {
	if st.db == nil {
		return nil, errors.New("database is not initialized")
	}

	if id == "" {
		return nil, errors.New("record id is empty")
	}

	list, err := st.RecordList(RecordQuery().
		SetID(id).
		SetLimit(1))

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

// RecordList returns a list of records
func (st *storeImplementation) RecordList(query RecordQueryInterface) ([]RecordInterface, error) {
	if st.db == nil {
		return nil, errors.New("database is not initialized")
	}

	q, columns, err := query.ToSelectDataset(st.dbDriverName, st.tableName)

	if err != nil {
		return []RecordInterface{}, err
	}

	sqlStr, sqlParams, errSql := q.Select(columns...).Prepared(true).ToSQL()

	if errSql != nil {
		return []RecordInterface{}, nil
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	modelMaps, err := database.SelectToMapString(database.Context(context.Background(), st.db), sqlStr, sqlParams...)

	if err != nil {
		return []RecordInterface{}, err
	}

	list := []RecordInterface{}

	lo.ForEach(modelMaps, func(modelMap map[string]string, index int) {
		model := NewRecordFromExistingData(modelMap)
		list = append(list, model)
	})

	return list, nil
}

func (store *storeImplementation) RecordSoftDelete(record RecordInterface) error {
	if record == nil {
		return errors.New("record is nil")
	}

	record.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RecordUpdate(record)
}

// RecordSoftDeleteByID soft deletes a record by ID
func (store *storeImplementation) RecordSoftDeleteByID(id string) error {
	if id == "" {
		return errors.New("record id is empty")
	}

	record, err := store.RecordFindByID(id)

	if err != nil {
		return err
	}

	if record == nil {
		return nil // Record does not exist, or is already soft deleted
	}

	return store.RecordSoftDelete(record)
}

// RecordUpdate updates a record
func (st *storeImplementation) RecordUpdate(record RecordInterface) error {
	if st.db == nil {
		return errors.New("database is not initialized")
	}

	if record == nil {
		return errors.New("record is nil")
	}

	if record.ID() == "" {
		return errors.New("record id is required")
	}

	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	dataChanged := record.DataChanged()

	delete(dataChanged, COLUMN_ID) // ID is not updateable

	if len(dataChanged) < 1 {
		return nil
	}

	sqlStr, params, errSql := goqu.Dialect(st.dbDriverName).
		Update(st.tableName).
		Prepared(true).
		Set(dataChanged).
		Where(goqu.C(COLUMN_ID).Eq(record.ID())).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr, params...)

	record.MarkAsNotDirty()

	return err
}
