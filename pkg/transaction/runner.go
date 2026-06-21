package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/gorm"
)

// ErrNestedTransaction reports that a transaction already exists in the context.
var ErrNestedTransaction = errors.New("nested transaction")

// Runner runs application work inside a transaction.
type Runner interface {
	// WithinTx runs fn inside a transaction.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// GormRunner implements Runner with GORM.
type GormRunner struct {
	db *gorm.DB // db stores the db value.
}

// New creates a GORM transaction runner.
func New(db *gorm.DB) GormRunner {
	return GormRunner{db: db}
}

// WithinTx runs fn inside a GORM transaction.
func (runner GormRunner) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := orm.Transaction(ctx); ok {
		return ErrNestedTransaction
	}

	tx := runner.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}

	if err := fn(orm.WithTransaction(ctx, tx)); err != nil {
		return rollback(tx, err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// rollback rolls back tx and joins rollback failures with cause.
func rollback(tx *gorm.DB, cause error) error {
	if err := tx.Rollback().Error; err != nil {
		return errors.Join(cause, fmt.Errorf("rollback transaction: %w", err))
	}
	return cause
}
