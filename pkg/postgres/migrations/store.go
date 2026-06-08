package migrations

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Record describes one stored migration history row.
type Record struct {
	// Version is the global migration version.
	Version int64

	// Name is the migration name.
	Name string

	// Direction is the applied direction.
	Direction Direction

	// Checksum is the applied up checksum.
	Checksum string

	// StartedAt is the execution start time.
	StartedAt time.Time

	// FinishedAt is the execution finish time.
	FinishedAt *time.Time

	// DurationMS is the execution duration in milliseconds.
	DurationMS *int64

	// Success reports whether the migration completed.
	Success bool

	// Error stores the failure message.
	Error string

	// Executor identifies the migration executor.
	Executor string

	// AppVersion stores the application version when known.
	AppVersion string

	// Dirty reports whether the database requires repair.
	Dirty bool
}

// Store persists schema migration history.
type Store struct {
	db *gorm.DB
}

// NewStore creates a migration history store.
func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

// Ensure creates the schema migration history table when missing.
func (store Store) Ensure(ctx context.Context) error {
	return store.db.WithContext(ctx).Exec(historyTableSQL(store.db.Dialector.Name())).Error
}

// Applied returns all recorded migration rows.
func (store Store) Applied(ctx context.Context) ([]Record, error) {
	var rows []historyRow
	err := store.db.WithContext(ctx).Raw("SELECT version, name, direction, checksum, started_at, finished_at, duration_ms, success, error, executor, app_version, dirty FROM gamehub_schema_migrations ORDER BY version ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, row.record())
	}
	return records, nil
}

// Start records migration execution start.
func (store Store) Start(ctx context.Context, migration Migration, executor string, appVersion string) error {
	row := historyRow{
		Version:    migration.Version,
		Name:       migration.Name,
		Direction:  string(DirectionUp),
		Checksum:   migration.Checksum,
		StartedAt:  time.Now().UTC(),
		Success:    false,
		Executor:   executor,
		AppVersion: appVersion,
		Dirty:      true,
	}
	return store.db.WithContext(ctx).Exec("INSERT INTO gamehub_schema_migrations(version, name, direction, checksum, started_at, success, error, executor, app_version, dirty) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", row.Version, row.Name, row.Direction, row.Checksum, row.StartedAt, row.Success, row.Error, row.Executor, row.AppVersion, row.Dirty).Error
}

// Succeed marks migration execution successful.
func (store Store) Succeed(ctx context.Context, migration Migration, started time.Time) error {
	finished := time.Now().UTC()
	duration := finished.Sub(started).Milliseconds()
	return store.db.WithContext(ctx).Exec("UPDATE gamehub_schema_migrations SET finished_at = ?, duration_ms = ?, success = ?, error = ?, dirty = ? WHERE version = ?", finished, duration, true, "", false, migration.Version).Error
}

// Fail marks migration execution failed.
func (store Store) Fail(ctx context.Context, migration Migration, started time.Time, runErr error) error {
	finished := time.Now().UTC()
	duration := finished.Sub(started).Milliseconds()
	return store.db.WithContext(ctx).Exec("UPDATE gamehub_schema_migrations SET finished_at = ?, duration_ms = ?, success = ?, error = ?, dirty = ? WHERE version = ?", finished, duration, false, runErr.Error(), true, migration.Version).Error
}

// Repair clears a dirty migration after manual repair.
func (store Store) Repair(ctx context.Context, version int64, checksum string, reason string) error {
	result := store.db.WithContext(ctx).Exec("UPDATE gamehub_schema_migrations SET checksum = ?, error = ?, dirty = ?, success = ? WHERE version = ? AND dirty = ?", checksum, "repaired: "+reason, false, true, version, true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("dirty migration %06d not found", version)
	}
	return nil
}

// Delete removes an applied migration record after a down migration succeeds.
func (store Store) Delete(ctx context.Context, version int64) error {
	return store.db.WithContext(ctx).Exec("DELETE FROM gamehub_schema_migrations WHERE version = ?", version).Error
}

// historyRow is the database representation of a migration record.
type historyRow struct {
	Version    int64
	Name       string
	Direction  string
	Checksum   string
	StartedAt  time.Time
	FinishedAt *time.Time
	DurationMS *int64
	Success    bool
	Error      string
	Executor   string
	AppVersion string
	Dirty      bool
}

// record maps row to Record.
func (row historyRow) record() Record {
	return Record{
		Version:    row.Version,
		Name:       row.Name,
		Direction:  Direction(row.Direction),
		Checksum:   row.Checksum,
		StartedAt:  row.StartedAt,
		FinishedAt: row.FinishedAt,
		DurationMS: row.DurationMS,
		Success:    row.Success,
		Error:      row.Error,
		Executor:   row.Executor,
		AppVersion: row.AppVersion,
		Dirty:      row.Dirty,
	}
}

// historyTableSQL returns schema history creation SQL.
func historyTableSQL(dialect string) string {
	if dialect == "sqlite" {
		return "CREATE TABLE IF NOT EXISTS gamehub_schema_migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL, direction TEXT NOT NULL, checksum TEXT NOT NULL, started_at DATETIME NOT NULL, finished_at DATETIME NULL, duration_ms INTEGER NULL, success BOOLEAN NOT NULL DEFAULT FALSE, error TEXT NULL, executor TEXT NOT NULL, app_version TEXT NULL, dirty BOOLEAN NOT NULL DEFAULT FALSE)"
	}
	return "CREATE TABLE IF NOT EXISTS gamehub_schema_migrations (version BIGINT PRIMARY KEY, name TEXT NOT NULL, direction TEXT NOT NULL, checksum TEXT NOT NULL, started_at TIMESTAMPTZ NOT NULL, finished_at TIMESTAMPTZ NULL, duration_ms BIGINT NULL, success BOOLEAN NOT NULL DEFAULT FALSE, error TEXT NULL, executor TEXT NOT NULL, app_version TEXT NULL, dirty BOOLEAN NOT NULL DEFAULT FALSE)"
}
