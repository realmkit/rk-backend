package orm

import (
	"context"

	"gorm.io/gorm"
)

// Store provides composable access to a GORM database handle.
type Store struct {
	db *gorm.DB // db stores the db value.
}

// NewStore creates a Store backed by db.
func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

// DB returns the active transaction from ctx when present, otherwise the base DB.
func (store Store) DB(ctx context.Context) *gorm.DB {
	if tx, ok := Transaction(ctx); ok {
		return tx
	}
	return store.db.WithContext(ctx)
}

// WithTx returns a Store backed directly by tx.
func (store Store) WithTx(tx *gorm.DB) Store {
	return Store{db: tx}
}

// transactionContextKey is the context key used for active GORM transactions.
type transactionContextKey struct{}

// WithTransaction returns a context carrying an active GORM transaction.
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionContextKey{}, tx)
}

// Transaction returns the active GORM transaction from ctx.
func Transaction(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(transactionContextKey{}).(*gorm.DB)
	return tx, ok && tx != nil
}
