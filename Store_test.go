// Package customstore_test provides black-box tests for the customstore package.
package customstore_test // Changed package name

import (
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	// Import the package we are testing
	"github.com/dracory/customstore" // Added import for the package itself

	"github.com/gouniverse/uid"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB remains a helper function within the test package
func InitDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		panic(err) // Panic is acceptable in test setup helpers
	}

	return db
}

func TestNewStore(t *testing.T) {
	db := InitDB("test_data_store_new.db")
	defer db.Close() // Good practice to close the DB

	// Test with valid options
	// Use customstore.NewStoreOptions and customstore.NewStore
	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_new",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err) // Use %v for errors
	}

	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	// Test with missing table name
	_, err = customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "",
		AutomigrateEnabled: true,
	})

	if err == nil {
		t.Fatalf("Expected error when creating store without table name, but got nil")
	}

	// Test with missing database
	_, err = customstore.NewStore(customstore.NewStoreOptions{
		DB:                 nil,
		TableName:          "data_new",
		AutomigrateEnabled: true,
	})

	if err == nil {
		t.Fatalf("Expected error when creating store without database, but got nil")
	}
}

func TestRecordCreate(t *testing.T) {
	db := InitDB("test_data_store_record_create.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_create",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	// Use customstore.NewRecord
	record := customstore.NewRecord("person")
	err = store.RecordCreate(record)

	if err != nil {
		t.Fatalf("Record could not be created: %v", err)
	}

	// Assuming default ID length is 32, check against that
	if len(record.ID()) != 32 {
		t.Fatalf("Expected Record ID length 32, but got %d (%s)", len(record.ID()), record.ID())
	}

	if record.CreatedAt() == "" {
		t.Fatalf("Record CreatedAt is empty")
	}

	if record.UpdatedAt() == "" {
		t.Fatalf("Record UpdatedAt is empty")
	}

	if record.IsSoftDeleted() {
		t.Fatalf("Record should not be soft deleted initially")
	}
}

func TestRecordFindByID(t *testing.T) {
	db := InitDB("test_data_store_record_find.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_find",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	record := customstore.NewRecord("person")
	// Use map[string]any for SetPayloadMap
	err = record.SetPayloadMap(map[string]any{
		"name": "Jon",
	})
	if err != nil {
		t.Fatalf("SetPayloadMap failed: %v", err)
	}

	err = store.RecordCreate(record)
	if err != nil {
		t.Fatalf("Record could not be created: %v", err)
	}

	// Check ID length again for consistency
	if len(record.ID()) != 32 {
		t.Fatalf("Expected Record ID length 32, but got %d (%s)", len(record.ID()), record.ID())
	}

	retrievedRecord, errFind := store.RecordFindByID(record.ID())
	if errFind != nil {
		t.Fatalf("RecordFindByID failed: %v", errFind)
	}

	if retrievedRecord == nil {
		t.Fatalf("Expected record to be found, but got nil")
	}

	// Compare payload string
	expectedPayload := `{"name":"Jon"}`
	if retrievedRecord.Payload() != expectedPayload {
		t.Fatalf("Expected payload %q, but got %q", expectedPayload, retrievedRecord.Payload())
	}

	// Test with non-existent ID
	nonExistentID := uid.HumanUid()
	retrievedRecord, errFind = store.RecordFindByID(nonExistentID)
	// Expecting NO error when record is not found, just a nil record
	if errFind != nil {
		t.Fatalf("RecordFindByID for non-existent ID failed unexpectedly: %v", errFind)
	}

	if retrievedRecord != nil {
		t.Fatalf("Expected nil when finding non-existent record, but got a record with ID %s", retrievedRecord.ID())
	}
}

func TestRecordUpdate(t *testing.T) {
	db := InitDB("test_data_store_record_update.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_update",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	record := customstore.NewRecord(`person`)
	initialPayload := map[string]any{
		`first_name`: `John`,
		`last_name`:  `Doe`,
	}
	err = record.SetPayloadMap(initialPayload)
	if err != nil {
		t.Fatalf("Initial SetPayloadMap failed: %v", err)
	}

	err = store.RecordCreate(record)
	if err != nil {
		t.Fatalf("Record could not be created: %v", err)
	}
	initialUpdatedAt := record.UpdatedAt() // Store initial timestamp

	retrievedRecord, errFind := store.RecordFindByID(record.ID())
	if errFind != nil {
		t.Fatalf("RecordFindByID failed: %v", errFind)
	}
	if retrievedRecord == nil {
		t.Fatalf("Expected record to be found, but got nil")
	}

	// Verify initial payload (consider JSON key order might vary)
	retrievedPayloadMap, err := retrievedRecord.PayloadMap()
	if err != nil {
		t.Fatalf("retrievedRecord.PayloadMap failed: %v", err)
	}
	if !reflect.DeepEqual(initialPayload, retrievedPayloadMap) {
		t.Fatalf("Initial payload mismatch. Expected %v, got %v", initialPayload, retrievedPayloadMap)
	}

	// Update the record
	updatedPayload := map[string]any{
		`first_name`: `Jane`,
		`last_name`:  `Smith`,
		`country`:    `GB`,
	}
	err = retrievedRecord.SetPayloadMap(updatedPayload)
	if err != nil {
		t.Fatalf("Update SetPayloadMap failed: %v", err)
	}

	err = store.RecordUpdate(retrievedRecord)
	if err != nil {
		t.Fatalf("Record could not be updated: %v", err)
	}

	// Retrieve again to verify update
	retrievedRecord2, errFind := store.RecordFindByID(record.ID())
	if errFind != nil {
		t.Fatalf("RecordFindByID after update failed: %v", errFind)
	}
	if retrievedRecord2 == nil {
		t.Fatalf("Expected record to be found after update, but got nil")
	}

	// Verify updated payload
	retrievedPayloadMap2, err := retrievedRecord2.PayloadMap()
	if err != nil {
		t.Fatalf("retrievedRecord2.PayloadMap failed: %v", err)
	}
	if !reflect.DeepEqual(updatedPayload, retrievedPayloadMap2) {
		t.Fatalf("Updated payload mismatch. Expected %v, got %v", updatedPayload, retrievedPayloadMap2)
	}

	// Verify UpdatedAt timestamp changed
	if retrievedRecord2.UpdatedAt() == initialUpdatedAt {
		t.Fatalf("Expected UpdatedAt timestamp to change after update, but it remained %s", initialUpdatedAt)
	}
}

func TestRecordDelete(t *testing.T) {
	db := InitDB("test_data_store_record_delete.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_delete",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	record := customstore.NewRecord("person")
	err = store.RecordCreate(record)
	if err != nil {
		t.Fatalf("Record could not be created: %v", err)
	}

	// Delete using the record object
	err = store.RecordDelete(record)
	if err != nil {
		t.Fatalf("RecordDelete failed: %v", err)
	}

	// Verify it's gone
	retrievedRecord, errFind := store.RecordFindByID(record.ID())
	if errFind != nil {
		t.Fatalf("RecordFindByID after delete failed unexpectedly: %v", errFind)
	}
	if retrievedRecord != nil {
		t.Fatalf("Expected record to be nil after delete, but found record with ID %s", retrievedRecord.ID())
	}

	// Test deleting a non-existent ID (should not error)
	err = store.RecordDeleteByID(uid.HumanUid())
	if err != nil {
		t.Fatalf("RecordDeleteByID for non-existent ID failed unexpectedly: %v", err)
	}
}

func TestRecordSoftDelete(t *testing.T) {
	db := InitDB("test_data_store_record_soft_delete.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_soft_delete",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	record := customstore.NewRecord("person")
	err = store.RecordCreate(record)
	if err != nil {
		t.Fatalf("Record could not be created: %v", err)
	}

	// Soft delete using the record object
	err = store.RecordSoftDelete(record)
	if err != nil {
		t.Fatalf("RecordSoftDelete failed: %v", err)
	}

	// Verify it's not found by default find
	retrievedRecord, errFind := store.RecordFindByID(record.ID())
	if errFind != nil {
		t.Fatalf("RecordFindByID after soft delete failed unexpectedly: %v", errFind)
	}
	if retrievedRecord != nil {
		t.Fatalf("Expected record to be nil after soft delete (default find), but found record with ID %s", retrievedRecord.ID())
	}

	// Test soft deleting a non-existent ID (should not error if find returns nil)
	err = store.RecordSoftDeleteByID(uid.HumanUid())
	if err != nil {
		t.Fatalf("RecordSoftDeleteByID for non-existent ID failed unexpectedly: %v", err)
	}

	// Test finding with soft deleted included
	// Use customstore.RecordQuery
	query := customstore.RecordQuery().SetSoftDeletedIncluded(true).SetID(record.ID())
	list, err := store.RecordList(query)
	if err != nil {
		t.Fatalf("RecordList with soft deleted included failed: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("Expected 1 record when including soft deleted, but got %d", len(list))
	}
	if list[0].ID() != record.ID() {
		t.Fatalf("Found wrong record ID. Expected %s, got %s", record.ID(), list[0].ID())
	}

	time.Sleep(1 * time.Second) // Wait for the soft delete timestamp to update (to be in the past)

	if !list[0].IsSoftDeleted() {
		t.Fatalf("Found record should be marked as soft deleted, but IsSoftDeleted returned false")
	}
}

func TestRecordList(t *testing.T) {
	db := InitDB("test_data_store_record_list.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_list",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	// Create records
	record1 := customstore.NewRecord("person")
	err = record1.SetPayloadMap(map[string]any{"name": "Jon"})
	if err != nil {
		t.Fatalf("SetPayloadMap record1 failed: %v", err)
	}
	err = store.RecordCreate(record1)
	if err != nil {
		t.Fatalf("RecordCreate record1 failed: %v", err)
	}

	record2 := customstore.NewRecord("person")
	err = record2.SetPayloadMap(map[string]any{"name": "Jane"})
	if err != nil {
		t.Fatalf("SetPayloadMap record2 failed: %v", err)
	}
	err = store.RecordCreate(record2)
	if err != nil {
		t.Fatalf("RecordCreate record2 failed: %v", err)
	}

	// List all records
	list, err := store.RecordList(customstore.RecordQuery())
	if err != nil {
		t.Fatalf("RecordList failed: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("Expected list length 2, but got %d", len(list))
	}

	// Check payloads (more robustly)
	foundNames := map[string]bool{}
	for _, rec := range list {
		payloadMap, err := rec.PayloadMap()
		if err != nil {
			t.Errorf("PayloadMap failed for record %s: %v", rec.ID(), err)
			continue
		}
		if name, ok := payloadMap["name"].(string); ok {
			foundNames[name] = true
		} else {
			t.Errorf("Payload for record %s does not contain a string 'name': %v", rec.ID(), payloadMap)
		}
	}

	if !foundNames["Jon"] {
		t.Errorf("Expected to find record with name 'Jon', but did not")
	}
	if !foundNames["Jane"] {
		t.Errorf("Expected to find record with name 'Jane', but did not")
	}
}

func TestRecordCount(t *testing.T) {
	db := InitDB("test_data_store_record_count.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_count",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	// Create records
	record1 := customstore.NewRecord("person")
	err = record1.SetPayloadMap(map[string]any{"name": "Jon"})
	if err != nil {
		t.Fatalf("SetPayloadMap record1 failed: %v", err)
	}
	err = store.RecordCreate(record1)
	if err != nil {
		t.Fatalf("RecordCreate record1 failed: %v", err)
	}

	record2 := customstore.NewRecord("person")
	err = record2.SetPayloadMap(map[string]any{"name": "Jane"})
	if err != nil {
		t.Fatalf("SetPayloadMap record2 failed: %v", err)
	}
	err = store.RecordCreate(record2)
	if err != nil {
		t.Fatalf("RecordCreate record2 failed: %v", err)
	}

	// Count all records
	count, err := store.RecordCount(customstore.RecordQuery())
	if err != nil {
		t.Fatalf("RecordCount failed: %v", err)
	}

	if count != 2 {
		t.Fatalf("Expected count 2, but got %d", count)
	}
}

func TestRecordQuery(t *testing.T) {
	db := InitDB("test_data_store_record_query.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_query",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	// Create records
	record1 := customstore.NewRecord("person")
	err = record1.SetPayloadMap(map[string]any{"name": "Jon"})
	if err != nil {
		t.Fatalf("SetPayloadMap record1 failed: %v", err)
	}
	err = store.RecordCreate(record1)
	if err != nil {
		t.Fatalf("RecordCreate record1 failed: %v", err)
	}

	record2 := customstore.NewRecord("company")
	err = record2.SetPayloadMap(map[string]any{"name": "Acme"})
	if err != nil {
		t.Fatalf("SetPayloadMap record2 failed: %v", err)
	}
	err = store.RecordCreate(record2)
	if err != nil {
		t.Fatalf("RecordCreate record2 failed: %v", err)
	}

	// Test with type
	query := customstore.RecordQuery().SetType("person")
	list, err := store.RecordList(query)
	if err != nil {
		t.Fatalf("RecordList with type 'person' failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected list length 1 for type 'person', but got %d", len(list))
	}
	payloadMap, _ := list[0].PayloadMap()
	if name, _ := payloadMap["name"].(string); name != "Jon" {
		t.Fatalf("Expected record name 'Jon', but got %q", name)
	}

	// Test with limit
	query = customstore.RecordQuery().SetLimit(1)
	list, err = store.RecordList(query)
	if err != nil {
		t.Fatalf("RecordList with limit 1 failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected list length 1 for limit 1, but got %d", len(list))
	}

	// Test with offset (needs limit)
	query = customstore.RecordQuery().SetOffset(1).SetLimit(1) // Ensure limit is set
	list, err = store.RecordList(query)
	if err != nil {
		t.Fatalf("RecordList with offset 1 failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Expected list length 1 for offset 1, but got %d", len(list))
	}

	// Test with order by (use exported constant)
	query = customstore.RecordQuery().SetOrderBy(customstore.COLUMN_CREATED_AT)
	list, err = store.RecordList(query)
	if err != nil {
		t.Fatalf("RecordList with order by failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("Expected list length 2 for order by, but got %d", len(list))
	}
	// Add more specific checks for order if needed, e.g., comparing CreatedAt timestamps
}

// --- Tests for Empty ID Handling ---

func TestRecordCreateWithEmptyID(t *testing.T) {
	db := InitDB("test_data_store_record_create_empty_id.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_create_empty_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	record := customstore.NewRecord("person")
	record.SetID("") // Explicitly set empty ID
	err = store.RecordCreate(record)

	if err == nil {
		t.Fatalf("Expected error when creating record with empty ID, but got nil")
	}
	// Check specific error message if desired, e.g., errors.Is(err, expectedError)
	expectedErrorMsg := "record ID is required"
	if err.Error() != expectedErrorMsg {
		t.Fatalf("Expected error message %q, but got %q", expectedErrorMsg, err.Error())
	}
}

func TestRecordUpdateWithEmptyID(t *testing.T) {
	db := InitDB("test_data_store_record_update_empty_id.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_update_empty_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	record := customstore.NewRecord("person")
	// Need to create it first to attempt an update
	err = store.RecordCreate(record)
	if err != nil {
		t.Fatalf("Setup: RecordCreate failed: %v", err)
	}

	// Now try to update with an empty ID (on a different instance to simulate error)
	recordToUpdate := customstore.NewRecord("person")
	recordToUpdate.SetID("")
	err = store.RecordUpdate(recordToUpdate)

	if err == nil {
		t.Fatalf("Expected error when updating record with empty ID, but got nil")
	}
	expectedErrorMsg := "record id is required"
	if err.Error() != expectedErrorMsg {
		t.Fatalf("Expected error message %q, but got %q", expectedErrorMsg, err.Error())
	}
}

func TestRecordDeleteWithEmptyID(t *testing.T) {
	db := InitDB("test_data_store_record_delete_empty_id.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_delete_empty_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	err = store.RecordDeleteByID("")

	if err == nil {
		t.Fatalf("Expected error when deleting record with empty ID, but got nil")
	}
	expectedErrorMsg := "record id is empty"
	if err.Error() != expectedErrorMsg {
		t.Fatalf("Expected error message %q, but got %q", expectedErrorMsg, err.Error())
	}
}

func TestRecordFindByIDWithEmptyID(t *testing.T) {
	db := InitDB("test_data_store_record_find_empty_id.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_find_empty_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	_, err = store.RecordFindByID("")

	if err == nil {
		t.Fatalf("Expected error when finding record with empty ID, but got nil")
	}
	expectedErrorMsg := "record id is empty"
	if err.Error() != expectedErrorMsg {
		t.Fatalf("Expected error message %q, but got %q", expectedErrorMsg, err.Error())
	}
}

func TestRecordSoftDeleteWithEmptyID(t *testing.T) {
	db := InitDB("test_data_store_record_soft_delete_empty_id.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_soft_delete_empty_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	err = store.RecordSoftDeleteByID("")

	if err == nil {
		t.Fatalf("Expected error when soft deleting record with empty ID, but got nil")
	}
	expectedErrorMsg := "record id is empty"
	if err.Error() != expectedErrorMsg {
		t.Fatalf("Expected error message %q, but got %q", expectedErrorMsg, err.Error())
	}
}

// --- Payload Search Test ---

func TestRecordQueryPayloadSearch(t *testing.T) {
	db := InitDB("test_data_store_record_query_payload_search.db")
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_query_payload_search",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}
	if store == nil {
		t.Fatalf("Store is nil after creation without error")
	}

	// Create records
	recordsData := []map[string]any{
		{"type": "person", "payload": map[string]any{"name": "Jon Doe", "country": "US", "status": "approved"}},
		{"type": "person", "payload": map[string]any{"name": "Jane Smith", "country": "GB", "status": "draft"}},
		{"type": "person", "payload": map[string]any{"name": "Tom Brown", "country": "US", "status": "approved"}},
		{"type": "company", "payload": map[string]any{"name": "Acme Corp", "country": "US", "status": "approved"}},
		{"type": "company", "payload": map[string]any{"name": "Beta Inc", "country": "GB", "status": "draft"}},
	}

	for i, data := range recordsData {
		rec := customstore.NewRecord(data["type"].(string))
		err = rec.SetPayloadMap(data["payload"].(map[string]any))
		if err != nil {
			t.Fatalf("SetPayloadMap record %d failed: %v", i+1, err)
		}
		err = store.RecordCreate(rec)
		if err != nil {
			t.Fatalf("RecordCreate record %d failed: %v", i+1, err)
		}
	}

	// Define test cases
	testCases := []struct {
		searchTerm    string
		expectedCount int
		expectedName  string // Optional: Check name of the first result if count is 1
	}{
		{"Jon", 1, "Jon Doe"},
		{`"country":"US"`, 3, ""}, // Expect 3 results with US country
		{"Jane", 1, "Jane Smith"},
		{"Acme", 1, "Acme Corp"},
		{"Corp", 1, "Acme Corp"},
		{"Smith", 1, "Jane Smith"},
		{"NonExistent", 0, ""},
		// Test multiple status search with OR condition
		{`"status":"approved"`, 3, ""}, // Should find all approved records
		{`"status":"draft"`, 2, ""},    // Should find both draft records
	}

	for _, tc := range testCases {
		t.Run("Search_"+tc.searchTerm, func(t *testing.T) {
			query := customstore.RecordQuery().AddPayloadSearch(tc.searchTerm)
			list, err := store.RecordList(query)
			if err != nil {
				t.Fatalf("RecordList with payload search %q failed: %v", tc.searchTerm, err)
			}

			if len(list) != tc.expectedCount {
				t.Fatalf("Expected list length %d for search %q, but got %d", tc.expectedCount, tc.searchTerm, len(list))
			}

			if tc.expectedCount == 1 && tc.expectedName != "" {
				payloadMap, err := list[0].PayloadMap()
				if err != nil {
					t.Fatalf("PayloadMap failed for result of search %q: %v", tc.searchTerm, err)
				}
				if name, ok := payloadMap["name"].(string); !ok || name != tc.expectedName {
					t.Fatalf("Expected first result name %q for search %q, but got %q (payload: %v)", tc.expectedName, tc.searchTerm, name, payloadMap)
				}
			}
		})
	}

	// Test multiple payload searches with OR condition
	t.Run("Search_Multiple_Status", func(t *testing.T) {
		query := customstore.RecordQuery().
			AddPayloadSearch(`"status":"approved"`).
			AddPayloadSearch(`"status":"draft"`)

		list, err := store.RecordList(query)
		if err != nil {
			t.Fatalf("RecordList with multiple payload search failed: %v", err)
		}

		if len(list) != 5 {
			t.Fatalf("Expected 5 records (all records with either approved or draft status), but got %d", len(list))
		}

		// Verify we got both approved and draft records
		statusCounts := map[string]int{"approved": 0, "draft": 0}
		for _, record := range list {
			payloadMap, err := record.PayloadMap()
			if err != nil {
				t.Fatalf("PayloadMap failed: %v", err)
			}
			if status, ok := payloadMap["status"].(string); ok {
				statusCounts[status]++
			}
		}

		if statusCounts["approved"] != 3 {
			t.Errorf("Expected 3 approved records, but got %d", statusCounts["approved"])
		}
		if statusCounts["draft"] != 2 {
			t.Errorf("Expected 2 draft records, but got %d", statusCounts["draft"])
		}
	})

	// Test NOT condition
	t.Run("Search_With_Not", func(t *testing.T) {
		query := customstore.RecordQuery().
			AddPayloadSearch(`"status":"approved"`).
			AddPayloadSearchNot(`"name":"Tom Brown"`)

		list, err := store.RecordList(query)
		if err != nil {
			t.Fatalf("RecordList with NOT condition failed: %v", err)
		}

		if len(list) != 2 {
			t.Fatalf("Expected 2 records (approved status but not Tom), but got %d", len(list))
		}

		// Verify we got the right records
		for _, record := range list {
			payloadMap, err := record.PayloadMap()
			if err != nil {
				t.Fatalf("PayloadMap failed: %v", err)
			}

			// Should be approved
			if status, ok := payloadMap["status"].(string); !ok || status != "approved" {
				t.Errorf("Expected status 'approved', but got %q", status)
			}

			// Should not be Tom
			if name, ok := payloadMap["name"].(string); ok && name == "Tom Brown" {
				t.Errorf("Found record with name 'Tom Brown' which should have been excluded")
			}
		}
	})
}
