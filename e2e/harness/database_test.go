package harness

import (
	"testing"
)

// TestNewSQLiteDatabaseRunsMigrations verifies the local fixture is migrated.
func TestNewSQLiteDatabaseRunsMigrations(t *testing.T) {
	database := NewSQLiteDatabase(t)
	sqlDB, err := database.DB.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	var count int64
	if err := database.DB.Table("realmkit_schema_migrations").Count(&count).Error; err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count == 0 {
		t.Fatalf("schema_migrations count = 0, want applied migrations")
	}
}
