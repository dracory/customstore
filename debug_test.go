package customstore_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestDriverRegistration(t *testing.T) {
	// Test different driver names
	driverNames := []string{"sqlite3", "sqlite"}

	for _, driverName := range driverNames {
		db, err := sql.Open(driverName, ":memory:")
		if err != nil {
			t.Logf("Failed to open database with driver '%s': %v", driverName, err)
			continue
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			t.Logf("Failed to ping database with driver '%s': %v", driverName, err)
			continue
		}

		t.Logf("SQLite driver works correctly with driver name: %s", driverName)
		return
	}

	t.Fatal("No working SQLite driver found")
}
