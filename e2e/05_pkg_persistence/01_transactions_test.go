// Package persistence_e2e verifies persistence infrastructure through real fixtures.
package persistence_e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/transaction"
)

// TestPersistenceTransactionsCommitAndRollback verifies transaction boundaries.
func TestPersistenceTransactionsCommitAndRollback(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start ecosystem with migrated database")
	ecosystem := harness.New(t)
	if err := ecosystem.Database.DB.AutoMigrate(&recordModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	runner := transaction.New(ecosystem.Database.DB)

	steps.Log("run transaction that rolls back on error")
	rollbackErr := errors.New("force rollback")
	err := runner.WithinTx(context.Background(), func(ctx context.Context) error {
		return errors.Join(createRecord(ecosystem.Database.Store, ctx, "rollback"), rollbackErr)
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("WithinTx() error = %v, want rollback error", err)
	}
	assertRecordCount(t, ecosystem, 0)

	steps.Log("run transaction that commits")
	if err := runner.WithinTx(context.Background(), func(ctx context.Context) error {
		return createRecord(ecosystem.Database.Store, ctx, "commit")
	}); err != nil {
		t.Fatalf("WithinTx() commit error = %v", err)
	}
	assertRecordCount(t, ecosystem, 1)
}

// TestPersistencePaginationNormalizesRequest verifies pagination defaults used by routes.
func TestPersistencePaginationNormalizesRequest(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("normalize empty request")
	page, err := pagination.New(pagination.Request{})
	if err != nil {
		t.Fatalf("pagination.New() error = %v", err)
	}
	if page.Limit != pagination.DefaultLimit {
		t.Fatalf("Limit = %d, want %d", page.Limit, pagination.DefaultLimit)
	}

	steps.Log("normalize oversized request")
	page, err = pagination.New(pagination.Request{Limit: pagination.MaxLimit + 100, Cursor: " next "})
	if err != nil {
		t.Fatalf("pagination.New() error = %v", err)
	}
	if page.Limit != pagination.MaxLimit || page.Cursor != "next" {
		t.Fatalf("page = %+v, want capped limit and trimmed cursor", page)
	}
}

// recordModel is a small persistence model owned by e2e tests.
type recordModel struct {
	orm.ID
	Name string `gorm:"column:name"`
}

// TableName returns the database table.
func (recordModel) TableName() string {
	return "e2e_records"
}

// createRecord writes one record through the ORM store.
func createRecord(store orm.Store, ctx context.Context, name string) error {
	return store.DB(ctx).Create(&recordModel{Name: name}).Error
}

// assertRecordCount verifies the number of records.
func assertRecordCount(t *testing.T, ecosystem *harness.Ecosystem, want int64) {
	t.Helper()
	var got int64
	if err := ecosystem.Database.DB.Table("e2e_records").Count(&got).Error; err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if got != want {
		t.Fatalf("record count = %d, want %d", got, want)
	}
}
