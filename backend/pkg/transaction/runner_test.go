package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/niflaot/gamehub/backend/pkg/orm"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// txModel is a persisted row used by transaction tests.
type txModel struct {
	ID   uint
	Name string
}

// TestWithinTxCommits verifies successful transaction work is committed.
func TestWithinTxCommits(t *testing.T) {
	db := openTransactionDB(t)
	runner := New(db)

	err := runner.WithinTx(context.Background(), func(ctx context.Context) error {
		return dbFromContext(t, ctx).Create(&txModel{Name: "committed"}).Error
	})
	if err != nil {
		t.Fatalf("WithinTx() error = %v", err)
	}

	var count int64
	if err := db.Model(&txModel{}).Where("name = ?", "committed").Count(&count).Error; err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want %d", count, 1)
	}
}

// TestWithinTxRollsBack verifies failed transaction work is rolled back.
func TestWithinTxRollsBack(t *testing.T) {
	db := openTransactionDB(t)
	runner := New(db)
	want := errors.New("fail")

	err := runner.WithinTx(context.Background(), func(ctx context.Context) error {
		if err := dbFromContext(t, ctx).Create(&txModel{Name: "rolled-back"}).Error; err != nil {
			return err
		}
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("WithinTx() error = %v, want %v", err, want)
	}

	var count int64
	if err := db.Model(&txModel{}).Where("name = ?", "rolled-back").Count(&count).Error; err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want %d", count, 0)
	}
}

// TestWithinTxRejectsNestedTransactions verifies nested transactions are explicit errors.
func TestWithinTxRejectsNestedTransactions(t *testing.T) {
	db := openTransactionDB(t)
	runner := New(db)
	tx := db.Begin()
	defer tx.Rollback()

	err := runner.WithinTx(orm.WithTransaction(context.Background(), tx), func(context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrNestedTransaction) {
		t.Fatalf("WithinTx() error = %v, want %v", err, ErrNestedTransaction)
	}
}

// TestWithinTxReturnsBeginErrors verifies begin failures are wrapped.
func TestWithinTxReturnsBeginErrors(t *testing.T) {
	db, mock, closeDB := openMockTransactionDB(t)
	defer closeDB()
	want := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(want)
	runner := New(db)

	err := runner.WithinTx(context.Background(), func(context.Context) error {
		t.Fatalf("transaction function called after begin failure")
		return nil
	})
	if err == nil {
		t.Fatalf("WithinTx() error = nil, want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

// TestWithinTxReturnsCommitErrors verifies commit failures are wrapped.
func TestWithinTxReturnsCommitErrors(t *testing.T) {
	db, mock, closeDB := openMockTransactionDB(t)
	defer closeDB()
	want := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(want)
	runner := New(db)

	err := runner.WithinTx(context.Background(), func(ctx context.Context) error {
		if _, ok := orm.Transaction(ctx); !ok {
			t.Fatalf("Transaction() ok = false, want true")
		}
		return nil
	})
	if err == nil {
		t.Fatalf("WithinTx() error = nil, want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

// TestRollbackJoinsRollbackErrors verifies rollback failures are joined with the cause.
func TestRollbackJoinsRollbackErrors(t *testing.T) {
	db, mock, closeDB := openMockTransactionDB(t)
	defer closeDB()
	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))
	tx := db.Begin()
	want := errors.New("cause")

	err := rollback(tx, want)
	if !errors.Is(err, want) {
		t.Fatalf("rollback() error = %v, want %v", err, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

// openTransactionDB opens an in-memory database for transaction tests.
func openTransactionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := db.AutoMigrate(&txModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

// openMockTransactionDB opens a sqlmock-backed GORM database.
func openMockTransactionDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	db, err := gorm.Open(postgresdriver.New(postgresdriver.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return db, mock, func() {
		_ = sqlDB.Close()
	}
}

// dbFromContext returns the transaction stored in ctx.
func dbFromContext(t *testing.T, ctx context.Context) *gorm.DB {
	t.Helper()
	db, ok := orm.Transaction(ctx)
	if !ok {
		t.Fatalf("Transaction() ok = false, want true")
	}
	return db
}
