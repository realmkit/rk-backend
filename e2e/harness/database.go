package harness

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database contains the migrated local database fixture.
type Database struct {
	// DB is the raw GORM database handle.
	DB *gorm.DB

	// Store is the project ORM wrapper over DB.
	Store orm.Store
}

// NewSQLiteDatabase creates an isolated migrated database for local e2e tests.
func NewSQLiteDatabase(t *testing.T) *Database {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(sqliteDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := postgres.Close(db); err != nil {
			t.Fatalf("postgres.Close() error = %v", err)
		}
	})

	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	return &Database{DB: db, Store: orm.NewStore(db)}
}

// sqliteDSN returns a shared in-memory SQLite DSN.
func sqliteDSN() string {
	return "file:" + uuid.NewString() + "?mode=memory&cache=shared"
}
