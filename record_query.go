package customstore

import (
	"errors"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dracory/sb"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

// RecordQuery shortcut for NewRecordQuery
func RecordQuery() RecordQueryInterface {
	return NewRecordQuery()
}

func NewRecordQuery() RecordQueryInterface {
	return &recordQueryImplementation{
		hasID:                 false,
		hasIDList:             false,
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

	// hasIDList is true if an ID list is set
	hasIDList bool

	// idList contains multiple IDs to match
	idList []string

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

	if o.IsIDListSet() && len(o.GetIDList()) == 0 {
		return errors.New("id list is required")
	}

	// sanitize empty strings out of the list
	filtered := lo.Filter(o.GetIDList(), func(id string, _ int) bool {
		return strings.TrimSpace(id) != ""
	})
	if o.IsIDListSet() && len(filtered) != len(o.GetIDList()) {
		return errors.New("id list contains empty strings")
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

	// Basic filters
	if o.IsIDSet() {
		q = q.Where(goqu.C(COLUMN_ID).Eq(o.GetID()))
	}
	if o.IsIDListSet() {
		// sanitize empty strings out of the list
		ids := []string{}
		for _, v := range o.GetIDList() {
			if strings.TrimSpace(v) != "" {
				ids = append(ids, v)
			}
		}
		if len(ids) > 0 {
			q = q.Where(goqu.C(COLUMN_ID).In(ids))
		}
	}

	// Payload conditions
	q = o.applyPayloadWhere(q)

	// Pagination and ordering
	q = o.applyPagination(q)
	q = o.applyOrderBy(q, sb.DESC)

	// Selected columns
	columns = o.selectedColumns()

	// Soft-delete and type constraints
	if o.IsTypeSet() {
		q = q.Where(goqu.C(COLUMN_RECORD_TYPE).Eq(o.GetType()))
	}
	return q.Where(o.softDeletedExpr()), columns, nil
}

// applyPayloadWhere applies payload include/exclude conditions.
func (o *recordQueryImplementation) applyPayloadWhere(q *goqu.SelectDataset) *goqu.SelectDataset {
	conds := []goqu.Expression{}

	if len(o.payloadSearch) > 0 {
		ors := make([]goqu.Expression, 0, len(o.payloadSearch))
		for _, v := range o.payloadSearch {
			ors = append(ors, goqu.I("payload").Like("%"+v+"%"))
		}
		conds = append(conds, goqu.Or(ors...))
	}

	if len(o.payloadSearchNot) > 0 {
		for _, v := range o.payloadSearchNot {
			conds = append(conds, goqu.I("payload").NotLike("%"+v+"%"))
		}
	}

	if len(conds) == 0 {
		return q
	}
	return q.Where(goqu.And(conds...))
}

// applyPagination applies limit/offset when not count-only.
func (o *recordQueryImplementation) applyPagination(q *goqu.SelectDataset) *goqu.SelectDataset {
	if o.IsOffsetSet() && !o.IsLimitSet() {
		o.SetLimit(10) // offset requires limit
	}
	if o.IsCountOnly() {
		return q
	}
	if o.IsLimitSet() {
		q = q.Limit(uint(o.GetLimit()))
	}
	if o.IsOffsetSet() {
		q = q.Offset(uint(o.GetOffset()))
	}
	return q
}

// applyOrderBy applies ordering if set.
func (o *recordQueryImplementation) applyOrderBy(q *goqu.SelectDataset, defaultOrder string) *goqu.SelectDataset {
	if !o.IsOrderBySet() {
		return q
	}
	if strings.EqualFold(defaultOrder, sb.ASC) {
		return q.Order(goqu.I(o.GetOrderBy()).Asc())
	}
	return q.Order(goqu.I(o.GetOrderBy()).Desc())
}

// selectedColumns returns any requested columns.
func (o *recordQueryImplementation) selectedColumns() []any {
	cols := []any{}
	for _, c := range o.GetColumns() {
		cols = append(cols, c)
	}
	return cols
}

// softDeletedExpr builds the soft-deleted filter expression.
func (o *recordQueryImplementation) softDeletedExpr() goqu.Expression {
	return goqu.C(COLUMN_SOFT_DELETED_AT).
		Gt(carbon.Now(carbon.UTC).ToDateTimeString())
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

// == ID LIST API ==
func (o *recordQueryImplementation) IsIDListSet() bool   { return o.hasIDList }
func (o *recordQueryImplementation) GetIDList() []string { return o.idList }
func (o *recordQueryImplementation) SetIDList(ids []string) RecordQueryInterface {
	o.hasIDList = true
	o.idList = ids
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
