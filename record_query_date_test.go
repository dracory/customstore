// Package customstore_test provides black-box tests for the customstore package.
package customstore_test

import (
	"testing"

	"github.com/dracory/customstore"

	_ "modernc.org/sqlite"
)

// seedRecordWithDate creates a record and overrides its created_at via raw SQL.
// RecordCreate overwrites created_at with carbon.Now(), so we must update it
// after creation to test date-range filtering.
func seedRecordWithDate(t *testing.T, store customstore.StoreInterface, recordType, createdAt string) customstore.RecordInterface {
	t.Helper()
	rec := customstore.NewRecord(recordType)
	if err := store.RecordCreate(rec); err != nil {
		t.Fatalf("RecordCreate failed: %v", err)
	}
	db := store.GetDB()
	if db == nil {
		t.Fatal("store DB is nil")
	}
	_, err := db.Exec("UPDATE data_record_date SET created_at = ? WHERE id = ?", createdAt, rec.ID())
	if err != nil {
		t.Fatalf("failed to override created_at: %v", err)
	}
	return rec
}

// TestRecordQueryCreatedAtGte verifies the created_at >= bound filters
// records correctly, including the boundary case where a record created
// exactly at the bound is included.
func TestRecordQueryCreatedAtGte(t *testing.T) {
	db := InitDB()
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_date",
		AutomigrateEnabled: true,
	})
	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}

	seedRecordWithDate(t, store, "contact", "2026-09-01 10:00:00")
	seedRecordWithDate(t, store, "contact", "2026-09-05 00:00:00")
	seedRecordWithDate(t, store, "contact", "2026-09-10 12:00:00")

	// date_from=2026-09-05 should include 2 records (09-05 and 09-10)
	list, err := store.RecordList(customstore.RecordQuery().
		SetType("contact").
		SetCreatedAtGte("2026-09-05 00:00:00").
		SetLimit(100))
	if err != nil {
		t.Fatalf("RecordList with CreatedAtGte failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 records with CreatedAtGte=2026-09-05, got %d", len(list))
	}
}

// TestRecordQueryCreatedAtLte verifies the created_at <= bound filters
// records correctly, including the boundary case.
func TestRecordQueryCreatedAtLte(t *testing.T) {
	db := InitDB()
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_date",
		AutomigrateEnabled: true,
	})
	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}

	seedRecordWithDate(t, store, "contact", "2026-09-01 10:00:00")
	seedRecordWithDate(t, store, "contact", "2026-09-05 00:00:00")
	seedRecordWithDate(t, store, "contact", "2026-09-10 12:00:00")

	// date_to=2026-09-05 should include 2 records (09-01 and 09-05)
	list, err := store.RecordList(customstore.RecordQuery().
		SetType("contact").
		SetCreatedAtLte("2026-09-05 23:59:59").
		SetLimit(100))
	if err != nil {
		t.Fatalf("RecordList with CreatedAtLte failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 records with CreatedAtLte=2026-09-05, got %d", len(list))
	}
}

// TestRecordQueryCreatedAtRange verifies both bounds work together.
func TestRecordQueryCreatedAtRange(t *testing.T) {
	db := InitDB()
	defer db.Close()

	store, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "data_record_date",
		AutomigrateEnabled: true,
	})
	if err != nil {
		t.Fatalf("Store could not be created: %v", err)
	}

	seedRecordWithDate(t, store, "contact", "2026-09-01 10:00:00") // before
	seedRecordWithDate(t, store, "contact", "2026-09-05 00:00:00") // start boundary
	seedRecordWithDate(t, store, "contact", "2026-09-10 12:00:00") // middle
	seedRecordWithDate(t, store, "contact", "2026-09-15 23:59:59") // end boundary
	seedRecordWithDate(t, store, "contact", "2026-09-20 10:00:00") // after

	// Range: 2026-09-05 to 2026-09-15 should include 3 records
	list, err := store.RecordList(customstore.RecordQuery().
		SetType("contact").
		SetCreatedAtGte("2026-09-05 00:00:00").
		SetCreatedAtLte("2026-09-15 23:59:59").
		SetLimit(100))
	if err != nil {
		t.Fatalf("RecordList with CreatedAt range failed: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 records in range 09-05..09-15, got %d", len(list))
	}
}

// TestRecordQueryCreatedAtEmptyString validates that SetCreatedAtGte("")
// clears the bound (does not set an empty-string bound that would fail).
func TestRecordQueryCreatedAtEmptyString(t *testing.T) {
	q := customstore.RecordQuery().
		SetCreatedAtGte("2026-09-05 00:00:00")

	if !q.IsCreatedAtGteSet() {
		t.Fatal("expected CreatedAtGte to be set after non-empty value")
	}
	if q.GetCreatedAtGte() != "2026-09-05 00:00:00" {
		t.Fatalf("expected CreatedAtGte=2026-09-05 00:00:00, got %q", q.GetCreatedAtGte())
	}

	// Empty string should clear the property, not set an empty bound
	q.SetCreatedAtGte("")
	if q.IsCreatedAtGteSet() {
		t.Fatal("expected CreatedAtGte to be cleared after empty string")
	}

	// Same for Lte
	q.SetCreatedAtLte("2026-09-15 23:59:59")
	if !q.IsCreatedAtLteSet() {
		t.Fatal("expected CreatedAtLte to be set after non-empty value")
	}
	q.SetCreatedAtLte("")
	if q.IsCreatedAtLteSet() {
		t.Fatal("expected CreatedAtLte to be cleared after empty string")
	}

	// Validate should pass with no bounds set
	if err := q.Validate(); err != nil {
		t.Fatalf("Validate should pass with no date bounds, got: %v", err)
	}
}
