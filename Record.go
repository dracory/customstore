package customstore

import (
	"encoding/json"

	"github.com/dracory/dataobject"
	"github.com/dracory/sb"
	"github.com/dracory/uid"
	"github.com/dromara/carbon/v2"
	"github.com/spf13/cast"
)

// ============================================================================
// == CLASS
// ============================================================================

type recordImplementation struct {
	dataobject.DataObject
}

// ============================================================================
// == CONSTRUCTORS
// ============================================================================

func NewRecord(recordType string, opts ...RecordOption) RecordInterface {
	record := recordImplementation{}
	record.SetID(uid.HumanUid())
	record.SetType(recordType)
	record.SetMemo("")
	record.SetMetas(map[string]string{})
	record.SetPayload("")
	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	record.SetSoftDeletedAt(sb.MAX_DATETIME)

	// Apply functional options, ignore errors to keep constructor signature simple
	for _, opt := range opts {
		_ = opt(&record)
	}

	return &record
}

func NewRecordFromExistingData(data map[string]string) RecordInterface {
	o := &recordImplementation{}
	o.Hydrate(data)
	return o
}

// ============================================================================
// == METHODS
// ============================================================================

func (o *recordImplementation) IsSoftDeleted() bool {
	return o.SoftDeletedAtCarbon().IsPast()
}

// ============================================================================
// == GETTERS AND SETTERS
// ============================================================================

func (o *recordImplementation) CreatedAt() string {
	return o.Get(COLUMN_CREATED_AT)
}

func (o *recordImplementation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.CreatedAt(), carbon.UTC)
}

func (o *recordImplementation) SetCreatedAt(createdAt string) {
	o.Set(COLUMN_CREATED_AT, createdAt)
}

func (o *recordImplementation) Type() string {
	return o.Get(COLUMN_RECORD_TYPE)
}

func (o *recordImplementation) SetType(recordType string) {
	o.Set(COLUMN_RECORD_TYPE, recordType)
}

func (o *recordImplementation) ID() string {
	return o.Get(COLUMN_ID)
}

func (o *recordImplementation) SetID(id string) {
	o.Set(COLUMN_ID, id)
}

func (o *recordImplementation) Memo() string {
	return o.Get(COLUMN_MEMO)
}

func (o *recordImplementation) SetMemo(memo string) {
	o.Set(COLUMN_MEMO, memo)
}

func (o *recordImplementation) Metas() (map[string]string, error) {
	metasStr := o.Get(COLUMN_METAS)

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
	o.Set(COLUMN_METAS, string(mapString))
	return nil
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
	return o.Get(COLUMN_PAYLOAD)
}

func (o *recordImplementation) SetPayload(payload string) {
	o.Set(COLUMN_PAYLOAD, payload)
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
	return o.Get(COLUMN_SOFT_DELETED_AT)
}

func (o *recordImplementation) SoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.SoftDeletedAt(), carbon.UTC)
}

func (o *recordImplementation) SetSoftDeletedAt(softDeletedAt string) {
	o.Set(COLUMN_SOFT_DELETED_AT, softDeletedAt)
}

func (o *recordImplementation) UpdatedAt() string {
	return o.Get(COLUMN_UPDATED_AT)
}

func (o *recordImplementation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(o.UpdatedAt(), carbon.UTC)
}

func (o *recordImplementation) SetUpdatedAt(updatedAt string) {
	o.Set(COLUMN_UPDATED_AT, updatedAt)
}
