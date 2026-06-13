package seeding

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Record describes one stored seed history row.
type Record struct {
	// Version is the global seed version.
	Version int64

	// Name is the seed name.
	Name string

	// Checksum is the applied checksum.
	Checksum string

	// StartedAt is the execution start time.
	StartedAt time.Time

	// FinishedAt is the execution finish time.
	FinishedAt *time.Time

	// DurationMS is the execution duration in milliseconds.
	DurationMS *int64

	// Success reports whether the seed completed.
	Success bool

	// Error stores the failure message.
	Error string

	// Executor identifies the seed executor.
	Executor string

	// AppVersion stores the application version when known.
	AppVersion string

	// Dirty reports whether the database requires repair.
	Dirty bool
}

// Store persists data seed history.
type Store struct {
	db *gorm.DB
}

// NewStore creates a data seed history store.
func NewStore(db *gorm.DB) Store {
	return Store{db: db}
}

// Ensure creates the data seed history table when missing.
func (store Store) Ensure(ctx context.Context) error {
	return store.db.WithContext(ctx).Exec(historyTableSQL(store.db.Dialector.Name())).Error
}

// Applied returns all recorded seed rows.
func (store Store) Applied(ctx context.Context) ([]Record, error) {
	var rows []historyRow
	err := store.db.WithContext(ctx).Raw(appliedSQL).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, row.record())
	}
	return records, nil
}

// Start records seed execution start.
func (store Store) Start(ctx context.Context, seed Seed, executor string, appVersion string) error {
	row := historyRow{
		Version:    seed.Version,
		Name:       seed.Name,
		Checksum:   seed.Checksum,
		StartedAt:  time.Now().UTC(),
		Success:    false,
		Executor:   executor,
		AppVersion: appVersion,
		Dirty:      true,
	}
	return store.db.WithContext(ctx).Exec(startSQL, row.startArgs()...).Error
}

// Succeed marks seed execution successful.
func (store Store) Succeed(ctx context.Context, seed Seed, started time.Time) error {
	finished := time.Now().UTC()
	duration := finished.Sub(started).Milliseconds()
	return store.db.WithContext(ctx).Exec(finishSQL, finished, duration, true, "", false, seed.Version).Error
}

// Fail marks seed execution failed.
func (store Store) Fail(ctx context.Context, seed Seed, started time.Time, runErr error) error {
	finished := time.Now().UTC()
	duration := finished.Sub(started).Milliseconds()
	return store.db.WithContext(ctx).Exec(finishSQL, finished, duration, false, runErr.Error(), true, seed.Version).Error
}

// Repair clears a dirty seed after manual repair.
func (store Store) Repair(ctx context.Context, version int64, checksum string, reason string) error {
	result := store.db.WithContext(ctx).Exec(repairSQL, checksum, "repaired: "+reason, false, true, version)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("seed %06d not found", version)
	}
	return nil
}

// historyRow is the database representation of a seed record.
type historyRow struct {
	Version    int64
	Name       string
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
	return Record(row)
}

// startArgs returns positional SQL arguments for a seed start record.
func (row historyRow) startArgs() []any {
	return []any{
		row.Version,
		row.Name,
		row.Checksum,
		row.StartedAt,
		row.Success,
		row.Error,
		row.Executor,
		row.AppVersion,
		row.Dirty,
	}
}

// historyTableSQL returns data seed history creation SQL.
func historyTableSQL(dialect string) string {
	if dialect == "sqlite" {
		return sqliteHistoryTableSQL
	}
	return postgresHistoryTableSQL
}

// Store SQL statements.
const (
	appliedSQL = `
SELECT version, name, checksum, started_at, finished_at, duration_ms,
	success, error, executor, app_version, dirty
FROM realmkit_data_seeds
ORDER BY version ASC`

	startSQL = `
INSERT INTO realmkit_data_seeds(
	version, name, checksum, started_at, success, error, executor, app_version, dirty
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	finishSQL = `
UPDATE realmkit_data_seeds
SET finished_at = ?, duration_ms = ?, success = ?, error = ?, dirty = ?
WHERE version = ?`

	repairSQL = `
UPDATE realmkit_data_seeds
SET checksum = ?, error = ?, dirty = ?, success = ?
WHERE version = ?`

	sqliteHistoryTableSQL = `
CREATE TABLE IF NOT EXISTS realmkit_data_seeds (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	checksum TEXT NOT NULL,
	started_at DATETIME NOT NULL,
	finished_at DATETIME NULL,
	duration_ms INTEGER NULL,
	success BOOLEAN NOT NULL DEFAULT FALSE,
	error TEXT NULL,
	executor TEXT NOT NULL,
	app_version TEXT NULL,
	dirty BOOLEAN NOT NULL DEFAULT FALSE
)`

	postgresHistoryTableSQL = `
CREATE TABLE IF NOT EXISTS realmkit_data_seeds (
	version BIGINT PRIMARY KEY,
	name TEXT NOT NULL,
	checksum TEXT NOT NULL,
	started_at TIMESTAMPTZ NOT NULL,
	finished_at TIMESTAMPTZ NULL,
	duration_ms BIGINT NULL,
	success BOOLEAN NOT NULL DEFAULT FALSE,
	error TEXT NULL,
	executor TEXT NOT NULL,
	app_version TEXT NULL,
	dirty BOOLEAN NOT NULL DEFAULT FALSE
)`
)
