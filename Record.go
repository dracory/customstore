package customstore

import (
	"encoding/json"

	"github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/soft_delete"
	neatuid "github.com/dracory/neat/support/uid"
	"github.com/dromara/carbon/v2"
	"github.com/spf13/cast"
)

// ============================================================================
// == INTERFACE
// ============================================================================

// RecordInterface represents an record for accessing the API
type RecordInterface interface {
	IsSoftDeleted() bool

	CreatedAt() string
	CreatedAtCarbon() *carbon.Carbon
	SetCreatedAt(createdAt string)

	ID() string
	SetID(id string)

	Type() string
	SetType(t string)

	Meta(name string) string
	SetMeta(name, value string) error

	Metas() (map[string]string, error)
	SetMetas(metas map[string]string) error
	UpsertMetas(metas map[string]string) error

	Memo() string
	SetMemo(memo string)

	Payload() string
	SetPayload(payload string)

	PayloadMap() (map[string]any, error)
	SetPayloadMap(payloadMap map[string]any) error
	PayloadMapKey(key string) (any, error)
	SetPayloadMapKey(key string, value any) error

	SoftDeletedAt() string
	SoftDeletedAtCarbon() *carbon.Carbon
	SetSoftDeletedAt(softDeletedAt string)

	UpdatedAt() string
	UpdatedAtCarbon() *carbon.Carbon
	SetUpdatedAt(updatedAt string)
}

// ============================================================================
// == TYPE
// ============================================================================

var _ RecordInterface = (*recordImplementation)(nil)

type recordImplementation struct {
	IDField        string `db:"id"`
	TypeField      string `db:"record_type"`
	PayloadField   string `db:"payload"`
	MetasField     string `db:"metas"`
	MemoField      string `db:"memo"`
	CreatedAtField orm.CreatedAt
	UpdatedAtField orm.UpdatedAt
	soft_delete.SoftDeletesMaxDate
}

// ============================================================================
// == CONSTRUCTORS
// ============================================================================

func NewRecord(recordType string, opts ...RecordOption) RecordInterface {
	record := &recordImplementation{}
	record.SetID(neatuid.GenerateShortID())
	record.SetType(recordType)
	record.SetMemo("")
	record.SetMetas(map[string]string{})
	record.SetPayload("")
	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	record.SetSoftDeletedAt(MAX_DATETIME)

	// Apply functional options, ignore errors to keep constructor signature simple
	for _, opt := range opts {
		_ = opt(record)
	}

	return record
}

func NewRecordFromExistingData(data map[string]string) RecordInterface {
	o := &recordImplementation{}
	if v, ok := data[COLUMN_ID]; ok {
		o.SetID(v)
	}
	if v, ok := data[COLUMN_RECORD_TYPE]; ok {
		o.SetType(v)
	}
	if v, ok := data[COLUMN_PAYLOAD]; ok {
		o.SetPayload(v)
	}
	if v, ok := data[COLUMN_METAS]; ok {
		o.SetMetasRaw(v)
	}
	if v, ok := data[COLUMN_MEMO]; ok {
		o.SetMemo(v)
	}
	if v, ok := data[COLUMN_CREATED_AT]; ok {
		o.SetCreatedAt(v)
	}
	if v, ok := data[COLUMN_UPDATED_AT]; ok {
		o.SetUpdatedAt(v)
	}
	if v, ok := data[COLUMN_SOFT_DELETED_AT]; ok {
		o.SetSoftDeletedAt(v)
	}
	return o
}

// ============================================================================
// == METHODS
// ============================================================================

func (o *recordImplementation) IsSoftDeleted() bool {
	return o.SoftDeletesMaxDate.SoftDeletedAt.Before(carbon.Now(carbon.UTC).StdTime())
}

// ============================================================================
// == GETTERS AND SETTERS
// ============================================================================

func (o *recordImplementation) CreatedAt() string {
	if o.CreatedAtField.CreatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.CreatedAtField.CreatedAt).ToDateTimeString()
}

func (o *recordImplementation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.CreatedAtField.CreatedAt)
}

func (o *recordImplementation) SetCreatedAt(createdAt string) {
	if createdAt == "" {
		return
	}
	o.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
}

func (o *recordImplementation) Type() string {
	return o.TypeField
}

func (o *recordImplementation) SetType(recordType string) {
	o.TypeField = recordType
}

func (o *recordImplementation) ID() string {
	return o.IDField
}

func (o *recordImplementation) SetID(id string) {
	o.IDField = id
}

func (o *recordImplementation) Memo() string {
	return o.MemoField
}

func (o *recordImplementation) SetMemo(memo string) {
	o.MemoField = memo
}

func (o *recordImplementation) Metas() (map[string]string, error) {
	metasStr := o.MetasField

	if metasStr == "" {
		metasStr = "{}"
	}

	var metas map[string]any
	err := json.Unmarshal([]byte(metasStr), &metas)
	if err != nil {
		return map[string]string{}, err
	}

	result := cast.ToStringMapString(metas)
	if result == nil {
		result = map[string]string{}
	}
	return result, nil
}

func (o *recordImplementation) Meta(name string) string {
	metas, err := o.Metas()

	if err != nil {
		return ""
	}

	if value, exists := metas[name]; exists {
		return value
	}

	return ""
}

func (o *recordImplementation) SetMeta(name, value string) error {
	return o.UpsertMetas(map[string]string{name: value})
}

// SetMetas stores metas as json string
// Warning: it overwrites any existing metas
func (o *recordImplementation) SetMetas(metas map[string]string) error {
	mapString, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	o.MetasField = string(mapString)
	return nil
}

// SetMetasRaw sets the metas field directly from a raw JSON string
func (o *recordImplementation) SetMetasRaw(metasStr string) {
	o.MetasField = metasStr
}

func (o *recordImplementation) UpsertMetas(metas map[string]string) error {
	currentMetas, err := o.Metas()

	if err != nil {
		return err
	}

	for k, v := range metas {
		currentMetas[k] = v
	}

	return o.SetMetas(currentMetas)
}

func (o *recordImplementation) Payload() string {
	return o.PayloadField
}

func (o *recordImplementation) SetPayload(payload string) {
	o.PayloadField = payload
}

func (r *recordImplementation) PayloadMap() (map[string]any, error) {
	data := make(map[string]any)

	if r.Payload() == "" {
		return data, nil
	}

	err := json.Unmarshal([]byte(r.Payload()), &data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (record *recordImplementation) SetPayloadMap(metas map[string]any) error {
	jsonBytes, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	jsonString := string(jsonBytes)
	record.SetPayload(jsonString)
	return nil
}

func (record *recordImplementation) PayloadMapKey(key string) (any, error) {
	data, err := record.PayloadMap()
	if err != nil {
		return nil, err
	}

	value, exists := data[key]
	if !exists {
		return nil, nil
	}

	return value, nil
}

func (record *recordImplementation) SetPayloadMapKey(key string, value any) error {
	data, err := record.PayloadMap()
	if err != nil {
		return err
	}

	data[key] = value

	return record.SetPayloadMap(data)
}

func (o *recordImplementation) SoftDeletedAt() string {
	if o.SoftDeletesMaxDate.SoftDeletedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.SoftDeletesMaxDate.SoftDeletedAt).ToDateTimeString()
}

func (o *recordImplementation) SoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.SoftDeletesMaxDate.SoftDeletedAt)
}

func (o *recordImplementation) SetSoftDeletedAt(softDeletedAt string) {
	if softDeletedAt == "" {
		return
	}
	o.SoftDeletesMaxDate.SoftDeletedAt = carbon.Parse(softDeletedAt, carbon.UTC).StdTime()
}

func (o *recordImplementation) UpdatedAt() string {
	if o.UpdatedAtField.UpdatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.UpdatedAtField.UpdatedAt).ToDateTimeString()
}

func (o *recordImplementation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.UpdatedAtField.UpdatedAt)
}

func (o *recordImplementation) SetUpdatedAt(updatedAt string) {
	if updatedAt == "" {
		return
	}
	o.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
}
