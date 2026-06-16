package customstore

import "errors"

// ============================================================================
// == INTERFACE
// ============================================================================

// RecordQueryInterface defines the interface for API record query operations
type RecordQueryInterface interface {
	Validate() error

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

// ============================================================================
// == TYPE
// ============================================================================

var _ RecordQueryInterface = (*recordQueryImplementation)(nil)

// ============================================================================
// == CONSTRUCTORS
// ============================================================================

// RecordQuery shortcut for NewRecordQuery
func RecordQuery() RecordQueryInterface {
	return NewRecordQuery()
}

func NewRecordQuery() RecordQueryInterface {
	return &recordQueryImplementation{
		properties: make(map[string]interface{}),
	}
}

// ============================================================================
// == CLASS
// ============================================================================

type recordQueryImplementation struct {
	properties map[string]interface{}
}

// ============================================================================
// == METHODS
// ============================================================================

func (o *recordQueryImplementation) Validate() error {
	if o.IsIDSet() && o.GetID() == "" {
		return errors.New("record query: id cannot be empty")
	}
	if o.IsIDListSet() && len(o.GetIDList()) == 0 {
		return errors.New("record query: id list cannot be empty")
	}
	if o.IsTypeSet() && o.GetType() == "" {
		return errors.New("record query: type cannot be empty")
	}
	if o.IsLimitSet() && o.GetLimit() < 0 {
		return errors.New("record query: limit cannot be negative")
	}
	if o.IsOffsetSet() && o.GetOffset() < 0 {
		return errors.New("record query: offset cannot be negative")
	}
	return nil
}

func (o *recordQueryImplementation) hasProperty(key string) bool {
	_, ok := o.properties[key]
	return ok
}

// == COLUMNS ==

func (o *recordQueryImplementation) SetColumns(columns []string) RecordQueryInterface {
	o.properties["columns"] = columns
	return o
}

func (o *recordQueryImplementation) GetColumns() []string {
	if v, ok := o.properties["columns"].([]string); ok {
		return v
	}
	return []string{}
}

// == COUNT ONLY ==

func (o *recordQueryImplementation) IsCountOnly() bool {
	return o.hasProperty("count_only")
}

func (o *recordQueryImplementation) SetCountOnly(countOnly bool) RecordQueryInterface {
	o.properties["count_only"] = countOnly
	return o
}

// == ID ==

func (o *recordQueryImplementation) IsIDSet() bool {
	return o.hasProperty("id")
}

func (o *recordQueryImplementation) GetID() string {
	return o.properties["id"].(string)
}

func (o *recordQueryImplementation) SetID(id string) RecordQueryInterface {
	if id == "" {
		delete(o.properties, "id")
	} else {
		o.properties["id"] = id
	}
	return o
}

// == ID LIST ==

func (o *recordQueryImplementation) IsIDListSet() bool {
	return o.hasProperty("id_list")
}

func (o *recordQueryImplementation) GetIDList() []string {
	return o.properties["id_list"].([]string)
}

func (o *recordQueryImplementation) SetIDList(ids []string) RecordQueryInterface {
	o.properties["id_list"] = ids
	return o
}

// == TYPE ==

func (o *recordQueryImplementation) IsTypeSet() bool {
	return o.hasProperty("type")
}

func (o *recordQueryImplementation) GetType() string {
	return o.properties["type"].(string)
}

func (o *recordQueryImplementation) SetType(recordType string) RecordQueryInterface {
	if recordType == "" {
		delete(o.properties, "type")
	} else {
		o.properties["type"] = recordType
	}
	return o
}

// == LIMIT ==

func (o *recordQueryImplementation) IsLimitSet() bool {
	return o.hasProperty("limit")
}

func (o *recordQueryImplementation) GetLimit() int {
	return o.properties["limit"].(int)
}

func (o *recordQueryImplementation) SetLimit(limit int) RecordQueryInterface {
	if limit < 0 {
		delete(o.properties, "limit")
	} else {
		o.properties["limit"] = limit
	}
	return o
}

// == OFFSET ==

func (o *recordQueryImplementation) IsOffsetSet() bool {
	return o.hasProperty("offset")
}

func (o *recordQueryImplementation) GetOffset() int {
	return o.properties["offset"].(int)
}

func (o *recordQueryImplementation) SetOffset(offset int) RecordQueryInterface {
	if offset < 0 {
		delete(o.properties, "offset")
	} else {
		o.properties["offset"] = offset
	}
	return o
}

// == ORDER BY ==

func (o *recordQueryImplementation) IsOrderBySet() bool {
	return o.hasProperty("order_by")
}

func (o *recordQueryImplementation) GetOrderBy() string {
	return o.properties["order_by"].(string)
}

func (o *recordQueryImplementation) SetOrderBy(orderBy string) RecordQueryInterface {
	if orderBy == "" {
		delete(o.properties, "order_by")
	} else {
		o.properties["order_by"] = orderBy
	}
	return o
}

// == SOFT DELETED INCLUDED ==

func (o *recordQueryImplementation) IsSoftDeletedIncluded() bool {
	return o.hasProperty("soft_deleted_included")
}

func (o *recordQueryImplementation) SetSoftDeletedIncluded(softDeletedIncluded bool) RecordQueryInterface {
	o.properties["soft_deleted_included"] = softDeletedIncluded
	return o
}

// == PAYLOAD SEARCH ==

func (o *recordQueryImplementation) AddPayloadSearch(needle string) RecordQueryInterface {
	if !o.hasProperty("payload_search") {
		o.properties["payload_search"] = []string{}
	}
	o.properties["payload_search"] = append(o.properties["payload_search"].([]string), needle)
	return o
}

func (o *recordQueryImplementation) GetPayloadSearch() []string {
	if v, ok := o.properties["payload_search"].([]string); ok {
		return v
	}
	return []string{}
}

// == PAYLOAD SEARCH NOT ==

func (o *recordQueryImplementation) AddPayloadSearchNot(needle string) RecordQueryInterface {
	if !o.hasProperty("payload_search_not") {
		o.properties["payload_search_not"] = []string{}
	}
	o.properties["payload_search_not"] = append(o.properties["payload_search_not"].([]string), needle)
	return o
}

func (o *recordQueryImplementation) GetPayloadSearchNot() []string {
	if v, ok := o.properties["payload_search_not"].([]string); ok {
		return v
	}
	return []string{}
}
