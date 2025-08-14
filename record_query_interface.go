package customstore

import (
	"github.com/doug-martin/goqu/v9"
)

// RecordQueryInterface defines the interface for API record query operations
type RecordQueryInterface interface {
	Validate() error
	ToSelectDataset(driver string, table string) (selectDataset *goqu.SelectDataset, columns []any, err error)

	IsSoftDeletedIncluded() bool
	SetSoftDeletedIncluded(softDeletedIncluded bool) RecordQueryInterface

	SetColumns(columns []string) RecordQueryInterface
	GetColumns() []string

	IsCountOnly() bool
	SetCountOnly(countOnly bool) RecordQueryInterface

	IsIDSet() bool
	GetID() string
	SetID(id string) RecordQueryInterface

	// Multiple IDs support
	IsIDListSet() bool
	GetIDList() []string
	SetIDList(ids []string) RecordQueryInterface

	IsTypeSet() bool
	GetType() string
	SetType(recordType string) RecordQueryInterface

	IsLimitSet() bool
	GetLimit() int
	SetLimit(limit int) RecordQueryInterface

	IsOffsetSet() bool
	GetOffset() int
	SetOffset(offset int) RecordQueryInterface

	IsOrderBySet() bool
	GetOrderBy() string
	SetOrderBy(orderBy string) RecordQueryInterface

	// Payload search methods
	AddPayloadSearch(needle string) RecordQueryInterface
	GetPayloadSearch() []string
	AddPayloadSearchNot(needle string) RecordQueryInterface
	GetPayloadSearchNot() []string
}
