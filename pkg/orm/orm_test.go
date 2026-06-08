package orm

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testModel composes ORM model helpers for tests.
type testModel struct {
	ID
	Timestamps
	SoftDelete
	Name string
}

// TestIDBeforeCreateAssignsUUID verifies UUIDs are generated before insert.
func TestIDBeforeCreateAssignsUUID(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	model := testModel{Name: "first"}
	if err := db.Create(&model).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if model.ID.ID == uuid.Nil {
		t.Fatalf("ID = %v, want generated UUID", model.ID.ID)
	}
	if model.CreatedAt.IsZero() {
		t.Fatalf("CreatedAt is zero")
	}
	if model.UpdatedAt.IsZero() {
		t.Fatalf("UpdatedAt is zero")
	}
}

// TestSoftDeleteHidesRows verifies soft-deleted rows are hidden by default.
func TestSoftDeleteHidesRows(t *testing.T) {
	db := openTestDB(t)
	if err := db.AutoMigrate(&testModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	model := testModel{Name: "deleted"}
	if err := db.Create(&model).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := db.Delete(&model).Error; err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var count int64
	if err := db.Model(&testModel{}).Where("id = ?", model.ID.ID).Count(&count).Error; err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want %d", count, 0)
	}
}

// TestStoreUsesContextTransaction verifies Store prefers transactions from context.
func TestStoreUsesContextTransaction(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	tx := db.Begin()
	defer tx.Rollback()

	got := store.DB(WithTransaction(context.Background(), tx))
	if got != tx {
		t.Fatalf("DB() = %p, want %p", got, tx)
	}
}

// TestStoreWithTxUsesProvidedTransaction verifies WithTx composes stores.
func TestStoreWithTxUsesProvidedTransaction(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	tx := db.Begin()
	defer tx.Rollback()

	got := store.WithTx(tx).DB(context.Background())
	if got.Statement.ConnPool != tx.Statement.ConnPool {
		t.Fatalf("WithTx().DB() did not use transaction connection")
	}
}

// TestTransactionRejectsMissingValues verifies Transaction ignores absent tx values.
func TestTransactionRejectsMissingValues(t *testing.T) {
	if _, ok := Transaction(context.Background()); ok {
		t.Fatalf("Transaction() ok = true, want false")
	}
}

// TestTranslateErrorMapsKnownErrors verifies repository error translation.
func TestTranslateErrorMapsKnownErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil", err: nil, want: nil},
		{name: "not_found", err: gorm.ErrRecordNotFound, want: ErrNotFound},
		{name: "conflict", err: &pgconn.PgError{Code: PostgresUniqueViolationCode}, want: ErrConflict},
		{name: "unavailable", err: context.DeadlineExceeded, want: ErrUnavailable},
		{name: "passthrough", err: errors.New("plain"), want: nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateError(tt.err)
			if tt.want == nil && got != tt.err {
				t.Fatalf("TranslateError() = %v, want %v", got, tt.err)
			}
			if tt.want != nil && !errors.Is(got, tt.want) {
				t.Fatalf("TranslateError() = %v, want wrapping %v", got, tt.want)
			}
		})
	}
}

// openTestDB opens an in-memory GORM database for ORM tests.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return db
}
