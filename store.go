package customstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dromara/carbon/v2"
)

// ============================================================================
// == INTERFACE
// ============================================================================

// StoreInterface defines a custom store
type StoreInterface interface {
	// MigrateDown drops the table
	MigrateDown(ctx context.Context, tx ...*sql.Tx) error

	// MigrateUp creates the table
	MigrateUp(ctx context.Context, tx ...*sql.Tx) error

	// EnableDebug - enables the debug option
	EnableDebug(debug bool)

	// GetDB returns the underlying *sql.DB
	GetDB() *sql.DB

	// RecordCount returns the count of records based on a query
	RecordCount(query RecordQueryInterface) (int64, error)

	// RecordCreate creates a new record
	RecordCreate(record RecordInterface) error

	// RecordDelete deletes a record
	RecordDelete(record RecordInterface) error

	// RecordDeleteByID deletes a record by ID
	RecordDeleteByID(id string) error

	// RecordFindByID finds a record by ID
	RecordFindByID(id string) (RecordInterface, error)

	// RecordList returns a list of records
	RecordList(query RecordQueryInterface) ([]RecordInterface, error)

	// RecordSoftDelete soft deletes a record
	RecordSoftDelete(record RecordInterface) error

	// RecordSoftDeleteByID soft deletes a record by ID
	RecordSoftDeleteByID(id string) error

	// RecordUpdate updates a record
	RecordUpdate(record RecordInterface) error
}

// ============================================================================
// == TYPE
// ============================================================================

var _ StoreInterface = (*storeImplementation)(nil)

// Store defines a custom store
type storeImplementation struct {
	tableName          string
	db                 *neat.Database
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
	if opts.DB == nil {
		return nil, errors.New("customstore store: DB is required")
	}

	if opts.TableName == "" {
		return nil, errors.New("customstore store: tableName is required")
	}

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	store := &storeImplementation{
		tableName:          opts.TableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 neatDB,
		debugEnabled:       opts.DebugEnabled,
		logger:             logger,
	}

	if store.automigrateEnabled {
		if err := store.MigrateUp(context.Background()); err != nil {
			return nil, err
		}
	}

	return store, nil
}

// ============================================================================
// == MIGRATE
// ============================================================================

// MigrateUp creates the table
func (st *storeImplementation) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if st.db.Schema().HasTable(st.tableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateUp: table already exists", "table", st.tableName)
		}
		return nil
	}

	err := st.db.Schema().Create(st.tableName, func(table contractsschema.Blueprint) {
		table.String(COLUMN_ID, 40)
		table.Primary(COLUMN_ID)
		table.String(COLUMN_RECORD_TYPE, 100)
		table.Text(COLUMN_PAYLOAD)
		table.Text(COLUMN_METAS)
		table.Text(COLUMN_MEMO)
		table.DateTime(COLUMN_CREATED_AT)
		table.DateTime(COLUMN_UPDATED_AT)
		table.DateTime(COLUMN_SOFT_DELETED_AT)
	})

	if err != nil {
		if st.debugEnabled {
			st.logger.Error("MigrateUp failed", "error", err)
		}
		return err
	}

	return nil
}

// MigrateDown drops the table
func (st *storeImplementation) MigrateDown(ctx context.Context, tx ...*sql.Tx) error {
	if !st.db.Schema().HasTable(st.tableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateDown: table does not exist", "table", st.tableName)
		}
		return nil
	}

	err := st.db.Schema().Drop(st.tableName)
	if err != nil {
		if st.debugEnabled {
			st.logger.Error("MigrateDown failed", "error", err)
		}
		return err
	}
	return nil
}

// ============================================================================
// == DEBUG
// ============================================================================

// EnableDebug - enables the debug option
func (st *storeImplementation) EnableDebug(debugEnabled bool) {
	st.debugEnabled = debugEnabled
	if debugEnabled {
		st.db.EnableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		st.db.DisableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}

// ============================================================================
// == DB
// ============================================================================

// GetDB returns the underlying *sql.DB.
func (st *storeImplementation) GetDB() *sql.DB {
	db, _ := st.db.DB()
	return db
}

// ============================================================================
// == RECORD CRUD
// ============================================================================

// RecordCount counts the number of records that match the query
func (st *storeImplementation) RecordCount(query RecordQueryInterface) (int64, error) {
	if st.db == nil {
		return 0, errors.New("database is not initialized")
	}

	q := st.buildQuery(query)

	var count int64
	err := q.Table(st.tableName).Count(&count)
	return count, err
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

	metas, err := record.Metas()
	if err != nil {
		return err
	}
	metasJSON, err := json.Marshal(metas)
	if err != nil {
		return err
	}

	row := map[string]any{
		COLUMN_ID:              record.ID(),
		COLUMN_RECORD_TYPE:     record.Type(),
		COLUMN_PAYLOAD:         record.Payload(),
		COLUMN_METAS:           string(metasJSON),
		COLUMN_MEMO:            record.Memo(),
		COLUMN_CREATED_AT:      record.CreatedAtCarbon().StdTime(),
		COLUMN_UPDATED_AT:      record.UpdatedAtCarbon().StdTime(),
		COLUMN_SOFT_DELETED_AT: record.SoftDeletedAtCarbon().StdTime(),
	}

	if st.debugEnabled {
		st.logger.Debug("Record create", "row", row)
	}

	return st.db.Query().Table(st.tableName).Create(row)
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

	_, err := st.db.Query().
		Table(st.tableName).
		Where(COLUMN_ID+" = ?", id).
		Delete()

	return err
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

	type recordRow struct {
		ID            string    `db:"id"`
		Type          string    `db:"record_type"`
		Payload       string    `db:"payload"`
		Metas         string    `db:"metas"`
		Memo          string    `db:"memo"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
		SoftDeletedAt time.Time `db:"soft_deleted_at"`
	}

	q := st.buildQuery(query)

	var rows []recordRow
	if err := q.Table(st.tableName).Get(&rows); err != nil {
		return []RecordInterface{}, err
	}

	list := make([]RecordInterface, 0, len(rows))
	for _, r := range rows {
		record := &recordImplementation{}
		record.SetID(r.ID)
		record.SetType(r.Type)
		record.SetPayload(r.Payload)
		record.SetMetasRaw(r.Metas)
		record.SetMemo(r.Memo)
		record.CreatedAtField.CreatedAt = r.CreatedAt
		record.UpdatedAtField.UpdatedAt = r.UpdatedAt
		record.SoftDeletesMaxDate.SoftDeletedAt = r.SoftDeletedAt
		list = append(list, record)
	}

	return list, nil
}

func (st *storeImplementation) RecordSoftDelete(record RecordInterface) error {
	if record == nil {
		return errors.New("record is nil")
	}

	return st.RecordSoftDeleteByID(record.ID())
}

// RecordSoftDeleteByID soft deletes a record by ID
func (st *storeImplementation) RecordSoftDeleteByID(id string) error {
	if id == "" {
		return errors.New("record id is empty")
	}

	row := map[string]any{
		COLUMN_SOFT_DELETED_AT: carbon.Now(carbon.UTC).StdTime(),
		COLUMN_UPDATED_AT:      carbon.Now(carbon.UTC).StdTime(),
	}

	_, err := st.db.Query().Table(st.tableName).Where(COLUMN_ID+" = ?", id).Update(row)
	return err
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

	metas, err := record.Metas()
	if err != nil {
		return err
	}
	metasJSON, err := json.Marshal(metas)
	if err != nil {
		return err
	}

	row := map[string]any{
		COLUMN_RECORD_TYPE: record.Type(),
		COLUMN_PAYLOAD:     record.Payload(),
		COLUMN_METAS:       string(metasJSON),
		COLUMN_MEMO:        record.Memo(),
		COLUMN_UPDATED_AT:  record.UpdatedAtCarbon().StdTime(),
	}

	if st.debugEnabled {
		st.logger.Debug("Record update", "row", row)
	}

	_, err = st.db.Query().Table(st.tableName).Where(COLUMN_ID+" = ?", record.ID()).Update(row)
	return err
}

// ============================================================================
// == QUERY BUILDER
// ============================================================================

// buildQuery builds a neat query from the record query interface.
func (st *storeImplementation) buildQuery(query RecordQueryInterface) contractsorm.Query {
	// Use Model() to enable neat's automatic soft delete handling via SoftDeletesMaxDate
	q := st.db.Query().Model(&recordImplementation{})

	if query == nil {
		return q
	}

	if query.IsIDSet() && query.GetID() != "" {
		q = q.Where(COLUMN_ID+" = ?", query.GetID())
	}

	if query.IsIDListSet() && len(query.GetIDList()) > 0 {
		idList := query.GetIDList()
		anyList := make([]any, len(idList))
		for i, v := range idList {
			anyList[i] = v
		}
		q = q.WhereIn(COLUMN_ID, anyList)
	}

	if query.IsTypeSet() && query.GetType() != "" {
		q = q.Where(COLUMN_RECORD_TYPE+" = ?", query.GetType())
	}

	if query.IsLimitSet() && query.GetLimit() > 0 {
		q = q.Limit(query.GetLimit())
	}

	if query.IsOffsetSet() && query.GetOffset() > 0 {
		q = q.Offset(query.GetOffset())
	}

	if query.IsOrderBySet() && query.GetOrderBy() != "" {
		q = q.OrderByDesc(query.GetOrderBy())
	}

	// Payload search (OR within positive searches, AND for negative)
	searchTerms := query.GetPayloadSearch()
	if len(searchTerms) > 0 {
		var searchQuery strings.Builder
		searchArgs := make([]any, 0, len(searchTerms))
		for i, needle := range searchTerms {
			if i > 0 {
				searchQuery.WriteString(" OR ")
			}
			searchQuery.WriteString(COLUMN_PAYLOAD + " LIKE ?")
			searchArgs = append(searchArgs, "%"+needle+"%")
		}
		q = q.Where("("+searchQuery.String()+")", searchArgs...)
	}
	for _, needle := range query.GetPayloadSearchNot() {
		q = q.Where(COLUMN_PAYLOAD+" NOT LIKE ?", "%"+needle+"%")
	}

	// Handle soft delete filtering via neat's automatic handling (SoftDeletesMaxDate)
	if query.IsSoftDeletedIncluded() {
		q = q.WithSoftDeleted()
	}

	return q
}
