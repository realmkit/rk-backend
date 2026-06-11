package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Locker manages migration mutual exclusion.
type Locker struct {
	db *gorm.DB
}

// NewLocker creates a migration locker.
func NewLocker(db *gorm.DB) Locker {
	return Locker{db: db}
}

// Lock acquires the migration advisory lock.
func (locker Locker) Lock(ctx context.Context) error {
	if locker.db.Dialector.Name() != "postgres" {
		return nil
	}
	return locker.db.WithContext(ctx).Exec("SELECT pg_advisory_lock(hashtext('realmkit_schema_migrations'))").Error
}

// Unlock releases the migration advisory lock.
func (locker Locker) Unlock(ctx context.Context) error {
	if locker.db.Dialector.Name() != "postgres" {
		return nil
	}
	return locker.db.WithContext(ctx).Exec("SELECT pg_advisory_unlock(hashtext('realmkit_schema_migrations'))").Error
}
