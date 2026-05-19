package customstore

import (
	"context"
	"database/sql"
)

// StoreInterface defines a custom store

type StoreInterface interface {
	// MigrateDown drops the table
	MigrateDown(ctx context.Context, tx ...*sql.Tx) error

	// MigrateUp creates the table
	MigrateUp(ctx context.Context, tx ...*sql.Tx) error

	// EnableDebug - enables the debug option
	EnableDebug(debug bool)

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
