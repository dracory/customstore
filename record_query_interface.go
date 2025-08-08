package customstore

import (
	"errors"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/sb"
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

// RecordQuery shortcut for NewRecordQuery
func RecordQuery() RecordQueryInterface {
	return NewRecordQuery()
}

func NewRecordQuery() RecordQueryInterface {
	return &recordQueryImplementation{
		hasID:                 false,
		isSoftDeletedIncluded: false,
		columns:               []string{},
		isCountOnly:           false,
		isLimitSet:            false,
		isOffsetSet:           false,
		isOrderBySet:          false,
		payloadSearch:         nil,
		payloadSearchNot:      nil,
	}
}

type recordQueryImplementation struct {
	// hasID is true if the ID is set, false otherwise
	hasID bool

	// id is the ID of the API record
	id string

	// isTypeSet is true if the record type is set, false otherwise
	isTypeSet bool

	// recordType is the record type of the API record
	recordType string

	// columns is the list of columns to select
	columns []string

	// isCountOnly is true if the query is for counting, false otherwise
	isCountOnly bool

	// isSoftDeletedIncluded is true if soft deleted records should be included, false otherwise
	isSoftDeletedIncluded bool

	isLimitSet bool

	// limit is the limit of the API record
	limit int

	isOffsetSet bool

	// offset is the offset of the API record
	offset int

	// isOrderBySet is true if the order by is set, false otherwise
	isOrderBySet bool

	// orderBy is the order by of the API record
	orderBy string

	// payloadSearch is the list of strings to search for in the payload
	payloadSearch []string

	// payloadSearchNot is the list of strings that should NOT be in the payload
	payloadSearchNot []string
}

func (o *recordQueryImplementation) Validate() error {
	if o.IsIDSet() && o.GetID() == "" {
		return errors.New("id is required")
	}

	if o.IsTypeSet() && o.GetType() == "" {
		return errors.New("type is required")
	}

	return nil
}

func (o *recordQueryImplementation) ToSelectDataset(driver string, table string) (selectDataset *goqu.SelectDataset, columns []any, err error) {
	if err := o.Validate(); err != nil {
		return nil, []any{}, err
	}

	q := goqu.Dialect(driver).From(table)

	if o.IsSoftDeletedIncluded() {
		return q, []any{}, nil // soft deleted sites requested specifically
	}

	// if o.IsCreatedAtGteSet() && o.IsCreatedAtLteSet() {
	// 	q = q.Where(
	// 		goqu.C(COLUMN_CREATED_AT).Gte(o.GetCreatedAtGte()),
	// 		goqu.C(COLUMN_CREATED_AT).Lte(o.GetCreatedAtLte()),
	// 	)
	// } else if o.IsCreatedAtGteSet() {
	// 	q = q.Where(goqu.C(COLUMN_CREATED_AT).Gte(o.GetCreatedAtGte()))
	// } else if o.IsCreatedAtLteSet() {
	// 	q = q.Where(goqu.C(COLUMN_CREATED_AT).Lte(o.GetCreatedAtLte()))
	// }

	if o.IsIDSet() {
		q = q.Where(goqu.C(COLUMN_ID).Eq(o.GetID()))
	}

	// if o.IsIDInSet() {
	// 	q = q.Where(goqu.C(COLUMN_ID).In(o.GetIDIn()))
	// }

	// if o.IsNameLikeSet() {
	// 	q = q.Where(goqu.C(COLUMN_NAME).Like("%" + o.GetNameLike() + "%"))
	// }

	// if o.IsStatusSet() {
	// 	q = q.Where(goqu.C(COLUMN_STATUS).Eq(o.GetStatus()))
	// }

	// if o.IsStatusInSet() {
	// 	q = q.Where(goqu.C(COLUMN_STATUS).In(o.GetStatusIn()))
	// }

	// Add payload search conditions
	conditions := []goqu.Expression{}

	if len(o.payloadSearch) > 0 {
		orConditions := []goqu.Expression{}
		for _, value := range o.payloadSearch {
			orConditions = append(orConditions, goqu.I("payload").Like("%"+value+"%"))
		}
		conditions = append(conditions, goqu.Or(orConditions...))
	}

	if len(o.payloadSearchNot) > 0 {
		for _, value := range o.payloadSearchNot {
			conditions = append(conditions, goqu.I("payload").NotLike("%"+value+"%"))
		}
	}

	if len(conditions) > 0 {
		q = q.Where(goqu.And(conditions...))
	}

	if o.IsOffsetSet() && !o.IsLimitSet() {
		o.SetLimit(10) // offset always requires limit to be set
	}

	if !o.IsCountOnly() {
		if o.IsLimitSet() {
			q = q.Limit(uint(o.GetLimit()))
		}

		if o.IsOffsetSet() {
			q = q.Offset(uint(o.GetOffset()))
		}
	}

	sortOrder := sb.DESC
	// if o.IsSortOrderSet() {
	// 	sortOrder = o.GetSortOrder()
	// }

	if o.IsOrderBySet() {
		if strings.EqualFold(sortOrder, sb.ASC) {
			q = q.Order(goqu.I(o.GetOrderBy()).Asc())
		} else {
			q = q.Order(goqu.I(o.GetOrderBy()).Desc())
		}
	}

	columns = []any{}

	for _, column := range o.GetColumns() {
		columns = append(columns, column)
	}

	if o.IsSoftDeletedIncluded() {
		return q, columns, nil // soft deleted sites requested specifically
	}

	softDeleted := goqu.C(COLUMN_SOFT_DELETED_AT).
		Gt(carbon.Now(carbon.UTC).ToDateTimeString())

	if o.IsTypeSet() {
		q = q.Where(goqu.C(COLUMN_RECORD_TYPE).Eq(o.GetType()))
	}

	return q.Where(softDeleted), columns, nil
}

func (o *recordQueryImplementation) SetColumns(columns []string) RecordQueryInterface {
	o.columns = columns
	return o
}

func (o *recordQueryImplementation) GetColumns() []string {
	return o.columns
}

func (o *recordQueryImplementation) IsCountOnly() bool {
	return o.isCountOnly
}

func (o *recordQueryImplementation) SetCountOnly(countOnly bool) RecordQueryInterface {
	o.isCountOnly = countOnly
	return o
}

func (o *recordQueryImplementation) IsIDSet() bool {
	return o.hasID
}

func (o *recordQueryImplementation) GetID() string {
	return o.id
}

func (o *recordQueryImplementation) SetID(id string) RecordQueryInterface {
	if id == "" {
		o.hasID = false
	} else {
		o.hasID = true
	}

	o.id = id

	return o
}

func (o *recordQueryImplementation) IsSoftDeletedIncluded() bool {
	return o.isSoftDeletedIncluded
}

func (o *recordQueryImplementation) SetSoftDeletedIncluded(softDeletedIncluded bool) RecordQueryInterface {
	o.isSoftDeletedIncluded = softDeletedIncluded
	return o
}

func (o *recordQueryImplementation) IsLimitSet() bool {
	return o.isLimitSet
}

func (o *recordQueryImplementation) GetLimit() int {
	return o.limit
}

func (o *recordQueryImplementation) SetLimit(limit int) RecordQueryInterface {
	o.isLimitSet = true
	o.limit = limit
	return o
}

func (o *recordQueryImplementation) IsOffsetSet() bool {
	return o.isOffsetSet
}

func (o *recordQueryImplementation) GetOffset() int {
	return o.offset
}

func (o *recordQueryImplementation) SetOffset(offset int) RecordQueryInterface {
	o.isOffsetSet = true
	o.offset = offset
	return o
}

func (o *recordQueryImplementation) IsOrderBySet() bool {
	return o.isOrderBySet
}

func (o *recordQueryImplementation) GetOrderBy() string {
	return o.orderBy
}

func (o *recordQueryImplementation) SetOrderBy(orderBy string) RecordQueryInterface {
	o.isOrderBySet = true
	o.orderBy = orderBy
	return o
}

func (o *recordQueryImplementation) IsTypeSet() bool {
	return o.isTypeSet
}

func (o *recordQueryImplementation) GetType() string {
	return o.recordType
}

func (o *recordQueryImplementation) SetType(recordType string) RecordQueryInterface {
	o.isTypeSet = true
	o.recordType = recordType
	return o
}

func (o *recordQueryImplementation) AddPayloadSearch(needle string) RecordQueryInterface {
	if o.payloadSearch == nil {
		o.payloadSearch = []string{}
	}
	o.payloadSearch = append(o.payloadSearch, needle)
	return o
}

func (o *recordQueryImplementation) GetPayloadSearch() []string {
	return o.payloadSearch
}

func (o *recordQueryImplementation) AddPayloadSearchNot(needle string) RecordQueryInterface {
	if o.payloadSearchNot == nil {
		o.payloadSearchNot = []string{}
	}
	o.payloadSearchNot = append(o.payloadSearchNot, needle)
	return o
}

func (o *recordQueryImplementation) GetPayloadSearchNot() []string {
	return o.payloadSearchNot
}
