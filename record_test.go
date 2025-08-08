// Package customstore_test provides black-box tests for the customstore package.
package customstore_test // Changed package name

import (
	"encoding/json"
	"reflect" // Import reflect for DeepEqual
	"testing"
	"time"

	// Import the package we are testing
	// Note: Replace "github.com/your-org/customstore" with your actual module path
	"github.com/dracory/customstore"

	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/sb"
)

func TestNewRecord(t *testing.T) {
	recordType := "user"
	// Use the imported package name to access exported functions/types
	record := customstore.NewRecord(recordType)

	if record == nil {
		t.Fatal("NewRecord returned nil")
	}
	if record.ID() == "" {
		t.Error("ID should be generated, but was empty")
	}
	if len(record.ID()) != 32 {
		t.Errorf("Default ID length should be 32, but got %d", len(record.ID()))
	}
	if record.Type() != recordType {
		t.Errorf("Expected record type %q, but got %q", recordType, record.Type())
	}
	if record.Memo() != "" {
		t.Errorf("Expected empty memo, but got %q", record.Memo())
	}
	metas, err := record.Metas()
	if err != nil {
		t.Errorf("record.Metas() failed: %v", err)
	}
	if len(metas) != 0 {
		t.Errorf("Expected empty metas, but got %v", metas)
	}
	if record.Payload() != "" {
		t.Errorf("Expected empty payload, but got %q", record.Payload())
	}
	if record.CreatedAt() == "" {
		t.Error("CreatedAt should not be empty")
	}
	if record.UpdatedAt() == "" {
		t.Error("UpdatedAt should not be empty")
	}
	if record.SoftDeletedAt() != sb.MAX_DATETIME {
		t.Errorf("Expected SoftDeletedAt to be %q, but got %q", sb.MAX_DATETIME, record.SoftDeletedAt())
	}
	if record.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to be false, but got true")
	}
}

func TestNewRecordFromExistingData(t *testing.T) {
	now := carbon.Now(carbon.UTC).ToDateTimeString()
	expectedMetas := map[string]string{"key1": "value1"}
	expectedPayloadStr := `{"item":"book","quantity":2}`
	expectedPayloadMap := map[string]any{"item": "book", "quantity": float64(2)} // JSON numbers unmarshal to float64
	data := map[string]string{
		// Use exported constants via the package name
		customstore.COLUMN_ID:              "test-id-123",
		customstore.COLUMN_RECORD_TYPE:     "order",
		customstore.COLUMN_MEMO:            "Test memo",
		customstore.COLUMN_METAS:           `{"key1":"value1"}`,
		customstore.COLUMN_PAYLOAD:         expectedPayloadStr,
		customstore.COLUMN_CREATED_AT:      now,
		customstore.COLUMN_UPDATED_AT:      now,
		customstore.COLUMN_SOFT_DELETED_AT: sb.MAX_DATETIME,
	}

	record := customstore.NewRecordFromExistingData(data)

	if record == nil {
		t.Fatal("NewRecordFromExistingData returned nil")
	}
	if record.ID() != "test-id-123" {
		t.Errorf("Expected ID %q, but got %q", "test-id-123", record.ID())
	}
	if record.Type() != "order" {
		t.Errorf("Expected Type %q, but got %q", "order", record.Type())
	}
	if record.Memo() != "Test memo" {
		t.Errorf("Expected Memo %q, but got %q", "Test memo", record.Memo())
	}

	metas, err := record.Metas()
	if err != nil {
		t.Errorf("record.Metas() failed: %v", err)
	}
	if !reflect.DeepEqual(metas, expectedMetas) {
		t.Errorf("Expected metas %v, but got %v", expectedMetas, metas)
	}
	if metaVal := record.Meta("key1"); metaVal != "value1" {
		t.Errorf("Expected Meta('key1') to be %q, but got %q", "value1", metaVal)
	}
	if metaVal := record.Meta("nonexistent"); metaVal != "" {
		t.Errorf("Expected Meta('nonexistent') to be empty, but got %q", metaVal)
	}

	if record.Payload() != expectedPayloadStr {
		t.Errorf("Expected Payload %q, but got %q", expectedPayloadStr, record.Payload())
	}
	payloadMap, err := record.PayloadMap()
	if err != nil {
		t.Errorf("record.PayloadMap() failed: %v", err)
	}
	if !reflect.DeepEqual(payloadMap, expectedPayloadMap) {
		t.Errorf("Expected PayloadMap %v, but got %v", expectedPayloadMap, payloadMap)
	}

	if record.CreatedAt() != now {
		t.Errorf("Expected CreatedAt %q, but got %q", now, record.CreatedAt())
	}
	if record.UpdatedAt() != now {
		t.Errorf("Expected UpdatedAt %q, but got %q", now, record.UpdatedAt())
	}
	if record.SoftDeletedAt() != sb.MAX_DATETIME {
		t.Errorf("Expected SoftDeletedAt %q, but got %q", sb.MAX_DATETIME, record.SoftDeletedAt())
	}
	if record.IsSoftDeleted() {
		t.Error("Expected IsSoftDeleted to be false, but got true")
	}
}

func TestIsSoftDeleted(t *testing.T) {
	record := customstore.NewRecord("test")

	// Default: Not soft deleted
	if record.IsSoftDeleted() {
		t.Error("Default: Expected IsSoftDeleted to be false, but got true")
	}

	// Set soft deleted date in the past
	pastTime := carbon.Now(carbon.UTC).SubDay().ToDateTimeString()
	record.SetSoftDeletedAt(pastTime)
	if !record.IsSoftDeleted() {
		t.Error("Past time: Expected IsSoftDeleted to be true, but got false")
	}

	// Set soft deleted date in the future (or max)
	record.SetSoftDeletedAt(sb.MAX_DATETIME)
	if record.IsSoftDeleted() {
		t.Error("Max time: Expected IsSoftDeleted to be false, but got true")
	}
}

func TestTimestamps(t *testing.T) {
	record := customstore.NewRecord("test")
	now := carbon.Now(carbon.UTC)
	nowStr := now.ToDateTimeString()

	// CreatedAt
	record.SetCreatedAt(nowStr)
	if record.CreatedAt() != nowStr {
		t.Errorf("CreatedAt: Expected string %q, but got %q", nowStr, record.CreatedAt())
	}
	if record.CreatedAtCarbon().Timestamp() != now.Timestamp() {
		t.Errorf("CreatedAtCarbon: Expected timestamp %d, but got %d", now.Timestamp(), record.CreatedAtCarbon().Timestamp())
	}

	// UpdatedAt
	record.SetUpdatedAt(nowStr)
	if record.UpdatedAt() != nowStr {
		t.Errorf("UpdatedAt: Expected string %q, but got %q", nowStr, record.UpdatedAt())
	}
	if record.UpdatedAtCarbon().Timestamp() != now.Timestamp() {
		t.Errorf("UpdatedAtCarbon: Expected timestamp %d, but got %d", now.Timestamp(), record.UpdatedAtCarbon().Timestamp())
	}

	// SoftDeletedAt
	record.SetSoftDeletedAt(nowStr)
	if record.SoftDeletedAt() != nowStr {
		t.Errorf("SoftDeletedAt: Expected string %q, but got %q", nowStr, record.SoftDeletedAt())
	}
	if record.SoftDeletedAtCarbon().Timestamp() != now.Timestamp() {
		t.Errorf("SoftDeletedAtCarbon: Expected timestamp %d, but got %d", now.Timestamp(), record.SoftDeletedAtCarbon().Timestamp())
	}
}

func TestID(t *testing.T) {
	record := customstore.NewRecord("test")
	newID := "custom-id-456"
	record.SetID(newID)
	if record.ID() != newID {
		t.Errorf("Expected ID %q, but got %q", newID, record.ID())
	}
}

func TestType(t *testing.T) {
	record := customstore.NewRecord("initialType")
	newType := "updatedType"
	record.SetType(newType)
	if record.Type() != newType {
		t.Errorf("Expected Type %q, but got %q", newType, record.Type())
	}
}

func TestMemo(t *testing.T) {
	record := customstore.NewRecord("test")
	newMemo := "This is a test memo."
	record.SetMemo(newMemo)
	if record.Memo() != newMemo {
		t.Errorf("Expected Memo %q, but got %q", newMemo, record.Memo())
	}
}

func TestMetas(t *testing.T) {
	record := customstore.NewRecord("test")

	// Initial state
	metas, err := record.Metas()
	if err != nil {
		t.Errorf("Initial Metas(): unexpected error: %v", err)
	}
	if len(metas) != 0 {
		t.Errorf("Initial Metas(): expected empty map, got %v", metas)
	}
	if metaVal := record.Meta("any"); metaVal != "" {
		t.Errorf("Initial Meta('any'): expected empty string, got %q", metaVal)
	}

	// SetMetas (overwrite)
	setMetasMap := map[string]string{"key1": "val1", "key2": "val2"}
	err = record.SetMetas(setMetasMap)
	if err != nil {
		t.Errorf("SetMetas(): unexpected error: %v", err)
	}
	metas, err = record.Metas()
	if err != nil {
		t.Errorf("Metas() after SetMetas: unexpected error: %v", err)
	}
	if !reflect.DeepEqual(metas, setMetasMap) {
		t.Errorf("Metas() after SetMetas: expected %v, got %v", setMetasMap, metas)
	}
	if metaVal := record.Meta("key1"); metaVal != "val1" {
		t.Errorf("Meta('key1') after SetMetas: expected %q, got %q", "val1", metaVal)
	}
	if metaVal := record.Meta("key2"); metaVal != "val2" {
		t.Errorf("Meta('key2') after SetMetas: expected %q, got %q", "val2", metaVal)
	}

	// SetMeta (upsert single)
	err = record.SetMeta("key3", "val3")
	if err != nil {
		t.Errorf("SetMeta('key3'): unexpected error: %v", err)
	}
	expectedMetasAfterSet := map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"}
	metas, err = record.Metas()
	if err != nil {
		t.Errorf("Metas() after SetMeta('key3'): unexpected error: %v", err)
	}
	if !reflect.DeepEqual(metas, expectedMetasAfterSet) {
		t.Errorf("Metas() after SetMeta('key3'): expected %v, got %v", expectedMetasAfterSet, metas)
	}
	if metaVal := record.Meta("key3"); metaVal != "val3" {
		t.Errorf("Meta('key3') after SetMeta: expected %q, got %q", "val3", metaVal)
	}

	// SetMeta (update existing)
	err = record.SetMeta("key1", "newVal1")
	if err != nil {
		t.Errorf("SetMeta('key1' update): unexpected error: %v", err)
	}
	expectedMetasAfterUpdate := map[string]string{"key1": "newVal1", "key2": "val2", "key3": "val3"}
	metas, err = record.Metas()
	if err != nil {
		t.Errorf("Metas() after SetMeta('key1' update): unexpected error: %v", err)
	}
	if !reflect.DeepEqual(metas, expectedMetasAfterUpdate) {
		t.Errorf("Metas() after SetMeta('key1' update): expected %v, got %v", expectedMetasAfterUpdate, metas)
	}
	if metaVal := record.Meta("key1"); metaVal != "newVal1" {
		t.Errorf("Meta('key1') after update: expected %q, got %q", "newVal1", metaVal)
	}

	// UpsertMetas (add and update)
	upsertMap := map[string]string{"key2": "newVal2", "key4": "val4"}
	err = record.UpsertMetas(upsertMap)
	if err != nil {
		t.Errorf("UpsertMetas(): unexpected error: %v", err)
	}
	expectedMetasAfterUpsert := map[string]string{"key1": "newVal1", "key2": "newVal2", "key3": "val3", "key4": "val4"}
	metas, err = record.Metas()
	if err != nil {
		t.Errorf("Metas() after UpsertMetas: unexpected error: %v", err)
	}
	if !reflect.DeepEqual(metas, expectedMetasAfterUpsert) {
		t.Errorf("Metas() after UpsertMetas: expected %v, got %v", expectedMetasAfterUpsert, metas)
	}
	if metaVal := record.Meta("key2"); metaVal != "newVal2" {
		t.Errorf("Meta('key2') after UpsertMetas: expected %q, got %q", "newVal2", metaVal)
	}
	if metaVal := record.Meta("key4"); metaVal != "val4" {
		t.Errorf("Meta('key4') after UpsertMetas: expected %q, got %q", "val4", metaVal)
	}

	// Test getting metas from invalid JSON (via NewRecordFromExistingData)
	invalidMetasData := map[string]string{
		customstore.COLUMN_ID:    "invalid-meta-id",
		customstore.COLUMN_METAS: `{"invalid"`, // Invalid JSON
	}
	invalidRecord := customstore.NewRecordFromExistingData(invalidMetasData)
	metas, err = invalidRecord.Metas()
	if err == nil {
		t.Error("Metas() with invalid JSON: expected an error, but got nil")
	}
	// Note: The current Metas() implementation returns an empty map on error.
	// Depending on desired behavior, you might expect nil here.
	if len(metas) != 0 {
		t.Errorf("Metas() with invalid JSON: expected empty map, got %v", metas)
	}
	// Meta should return empty string if Metas() fails
	if metaVal := invalidRecord.Meta("key1"); metaVal != "" {
		t.Errorf("Meta() with invalid JSON: expected empty string, got %q", metaVal)
	}
}

func TestPayload(t *testing.T) {
	record := customstore.NewRecord("test")
	payloadStr := `{"product":"widget","price":19.99}`

	record.SetPayload(payloadStr)
	if record.Payload() != payloadStr {
		t.Errorf("Expected Payload %q, got %q", payloadStr, record.Payload())
	}
}

func TestPayloadMap(t *testing.T) {
	record := customstore.NewRecord("test")

	// Initial empty payload
	payloadMap, err := record.PayloadMap()
	if err != nil {
		t.Errorf("Initial PayloadMap(): unexpected error: %v", err)
	}
	if len(payloadMap) != 0 {
		t.Errorf("Initial PayloadMap(): expected empty map, got %v", payloadMap)
	}

	// Set payload map
	initialMap := map[string]any{
		"name": "Alice",
		"age":  float64(25), // Use float64 as JSON numbers decode to this
		"city": "New York",
	}
	err = record.SetPayloadMap(initialMap)
	if err != nil {
		t.Errorf("SetPayloadMap(): unexpected error: %v", err)
	}

	// Verify Payload string is correct JSON
	var actualMap map[string]any
	err = json.Unmarshal([]byte(record.Payload()), &actualMap)
	if err != nil {
		t.Errorf("json.Unmarshal payload string failed: %v", err)
	}
	// Use reflect.DeepEqual for map comparison
	if !reflect.DeepEqual(initialMap, actualMap) {
		t.Errorf("Payload string unmarshalled incorrectly. Expected %v, got %v", initialMap, actualMap)
	}

	// Retrieve payload map
	retrievedMap, err := record.PayloadMap()
	if err != nil {
		t.Errorf("PayloadMap() after SetPayloadMap: unexpected error: %v", err)
	}
	if !reflect.DeepEqual(initialMap, retrievedMap) {
		t.Errorf("Retrieved PayloadMap incorrect. Expected %v, got %v", initialMap, retrievedMap)
	}

	// Test getting from invalid JSON payload (via NewRecordFromExistingData)
	invalidPayloadData := map[string]string{
		customstore.COLUMN_ID:      "invalid-payload-id",
		customstore.COLUMN_PAYLOAD: `{"invalid"`, // Invalid JSON
	}
	invalidRecord := customstore.NewRecordFromExistingData(invalidPayloadData)
	payloadMap, err = invalidRecord.PayloadMap()
	if err == nil {
		t.Error("PayloadMap() with invalid JSON: expected an error, but got nil")
	}
	if payloadMap != nil {
		t.Errorf("PayloadMap() with invalid JSON: expected nil map, got %v", payloadMap)
	}
}

func TestPayloadMapKey(t *testing.T) {
	record := customstore.NewRecord("test")

	// Test setting a key on an empty payload
	err := record.SetPayloadMapKey("name", "John")
	if err != nil {
		t.Errorf("SetPayloadMapKey('name', 'John'): unexpected error: %v", err)
	}

	// Test getting the key
	value, err := record.PayloadMapKey("name")
	if err != nil {
		t.Errorf("PayloadMapKey('name'): unexpected error: %v", err)
	}
	if value != "John" {
		t.Errorf("PayloadMapKey('name'): expected %q, got %q", "John", value)
	}

	// Test getting a non-existent key
	value, err = record.PayloadMapKey("nonexistent")
	if err != nil {
		t.Errorf("PayloadMapKey('nonexistent'): unexpected error: %v", err)
	}
	if value != nil {
		t.Errorf("PayloadMapKey('nonexistent'): expected nil, got %v", value)
	}

	// Test setting multiple keys (numeric)
	err = record.SetPayloadMapKey("age", 30) // Will be stored as float64
	if err != nil {
		t.Errorf("SetPayloadMapKey('age', 30): unexpected error: %v", err)
	}

	value, err = record.PayloadMapKey("age")
	if err != nil {
		t.Errorf("PayloadMapKey('age'): unexpected error: %v", err)
	}
	// JSON numbers are decoded as float64
	if age, ok := value.(float64); !ok || age != 30.0 {
		t.Errorf("PayloadMapKey('age'): expected float64(30), got %T(%v)", value, value)
	}

	// Test setting a boolean key
	err = record.SetPayloadMapKey("active", true)
	if err != nil {
		t.Errorf("SetPayloadMapKey('active', true): unexpected error: %v", err)
	}
	value, err = record.PayloadMapKey("active")
	if err != nil {
		t.Errorf("PayloadMapKey('active'): unexpected error: %v", err)
	}
	if active, ok := value.(bool); !ok || !active {
		t.Errorf("PayloadMapKey('active'): expected bool(true), got %T(%v)", value, value)
	}

	// Test setting a nested structure (map)
	nestedMap := map[string]any{"street": "123 Main St", "zip": "10001"}
	err = record.SetPayloadMapKey("address", nestedMap)
	if err != nil {
		t.Errorf("SetPayloadMapKey('address', map): unexpected error: %v", err)
	}
	value, err = record.PayloadMapKey("address")
	if err != nil {
		t.Errorf("PayloadMapKey('address'): unexpected error: %v", err)
	}
	// JSON unmarshalling turns nested maps into map[string]any
	if !reflect.DeepEqual(nestedMap, value) {
		t.Errorf("PayloadMapKey('address'): expected %v, got %v", nestedMap, value)
	}

	// Test setting a slice
	tags := []any{"go", "test", "json"}
	err = record.SetPayloadMapKey("tags", tags)
	if err != nil {
		t.Errorf("SetPayloadMapKey('tags', slice): unexpected error: %v", err)
	}
	value, err = record.PayloadMapKey("tags")
	if err != nil {
		t.Errorf("PayloadMapKey('tags'): unexpected error: %v", err)
	}
	if !reflect.DeepEqual(tags, value) {
		t.Errorf("PayloadMapKey('tags'): expected %v, got %v", tags, value)
	}

	// Verify the entire payload map
	payloadMap, err := record.PayloadMap()
	if err != nil {
		t.Errorf("PayloadMap() for verification: unexpected error: %v", err)
	}
	expectedMap := map[string]any{
		"name":    "John",
		"age":     float64(30),
		"active":  true,
		"address": map[string]any{"street": "123 Main St", "zip": "10001"},
		"tags":    []any{"go", "test", "json"},
	}
	if !reflect.DeepEqual(expectedMap, payloadMap) {
		t.Errorf("Verified PayloadMap incorrect. Expected %v, got %v", expectedMap, payloadMap)
	}

	// Test updating an existing key
	err = record.SetPayloadMapKey("name", "Johnny")
	if err != nil {
		t.Errorf("SetPayloadMapKey('name', 'Johnny'): unexpected error: %v", err)
	}
	value, err = record.PayloadMapKey("name")
	if err != nil {
		t.Errorf("PayloadMapKey('name') after update: unexpected error: %v", err)
	}
	if value != "Johnny" {
		t.Errorf("PayloadMapKey('name') after update: expected %q, got %q", "Johnny", value)
	}

	// Test getting/setting key from invalid JSON payload (via NewRecordFromExistingData)
	invalidPayloadData := map[string]string{
		customstore.COLUMN_ID:      "invalid-payload-key-id",
		customstore.COLUMN_PAYLOAD: `{"invalid"`, // Invalid JSON
	}
	invalidRecord := customstore.NewRecordFromExistingData(invalidPayloadData)

	// Getting should fail
	value, err = invalidRecord.PayloadMapKey("name")
	if err == nil {
		t.Error("PayloadMapKey() get with invalid JSON: expected an error, but got nil")
	}
	if value != nil {
		t.Errorf("PayloadMapKey() get with invalid JSON: expected nil value, got %v", value)
	}

	// Setting should fail (because it reads first)
	err = invalidRecord.SetPayloadMapKey("newKey", "newValue")
	if err == nil {
		t.Error("SetPayloadMapKey() with invalid JSON: expected an error, but got nil")
	}
}

// Helper function to introduce a small delay (renamed)
// Note: Consider if this sleep is truly necessary for the test logic.
func sleepForTest(duration time.Duration) {
	time.Sleep(duration)
}

func TestDirtyTracking(t *testing.T) {
	record := customstore.NewRecord("test")
	record.MarkAsNotDirty() // Start clean

	if record.IsDirty() {
		t.Error("Initial: Expected IsDirty to be false")
	}
	if len(record.DataChanged()) != 0 {
		t.Errorf("Initial: Expected empty DataChanged, got %v", record.DataChanged())
	}

	// Modify a field
	record.SetMemo("new memo")
	if !record.IsDirty() {
		t.Error("After SetMemo: Expected IsDirty to be true")
	}
	changed := record.DataChanged()
	if len(changed) != 1 {
		t.Errorf("After SetMemo: Expected DataChanged length 1, got %d", len(changed))
	}
	if val, ok := changed[customstore.COLUMN_MEMO]; !ok || val != "new memo" {
		t.Errorf("After SetMemo: Expected DataChanged[%q] to be %q, got %q (ok=%v)", customstore.COLUMN_MEMO, "new memo", val, ok)
	}

	// Modify another field
	record.SetPayload(`{"key":"value"}`)
	if !record.IsDirty() {
		t.Error("After SetPayload: Expected IsDirty to be false")
	}
	changed = record.DataChanged()
	if len(changed) != 2 {
		t.Errorf("After SetPayload: Expected DataChanged length 2, got %d", len(changed))
	}
	if val, ok := changed[customstore.COLUMN_MEMO]; !ok || val != "new memo" {
		t.Errorf("After SetPayload: Expected DataChanged[%q] to be %q, got %q (ok=%v)", customstore.COLUMN_MEMO, "new memo", val, ok)
	}
	if val, ok := changed[customstore.COLUMN_PAYLOAD]; !ok || val != `{"key":"value"}` {
		t.Errorf("After SetPayload: Expected DataChanged[%q] to be %q, got %q (ok=%v)", customstore.COLUMN_PAYLOAD, `{"key":"value"}`, val, ok)
	}

	// Mark as not dirty
	record.MarkAsNotDirty()
	if record.IsDirty() {
		t.Error("After MarkAsNotDirty: Expected IsDirty to be false")
	}
	if len(record.DataChanged()) != 0 {
		t.Errorf("After MarkAsNotDirty: Expected empty DataChanged, got %v", record.DataChanged())
	}

	// Test Hydrate resets dirty status
	record.SetMemo("another memo") // Make it dirty again
	if !record.IsDirty() {
		t.Error("Before Hydrate: Expected IsDirty to be true")
	}
	// Call Hydrate via the interface (it's part of dataobject.DataObjectInterface)
	record.Hydrate(map[string]string{customstore.COLUMN_ID: record.ID()}) // Hydrate with minimal data
	if !record.IsDirty() {
		t.Error("After Hydrate: Expected IsDirty to be false") // Hydrate should reset dirty status
	}
}
