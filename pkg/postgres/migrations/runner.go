package migrations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrDirty reports that the database has a dirty migration.
var ErrDirty = errors.New("database migration state is dirty")

// ErrChecksumChanged reports that an applied migration checksum changed.
var ErrChecksumChanged = errors.New("migration checksum changed")

// Runner applies and validates global schema migrations.
type Runner struct {
	db         *gorm.DB
	source     Source
	store      Store
	locker     Locker
	log        *zap.Logger
	executor   string
	appVersion string
}

// Option configures a Runner.
type Option func(*Runner)

// WithLogger configures migration logging.
func WithLogger(log *zap.Logger) Option {
	return func(runner *Runner) {
		runner.log = log
	}
}

// WithExecutor configures the migration executor name.
func WithExecutor(executor string) Option {
	return func(runner *Runner) {
		runner.executor = executor
	}
}

// WithAppVersion configures the application version written to history.
func WithAppVersion(version string) Option {
	return func(runner *Runner) {
		runner.appVersion = version
	}
}

// NewRunner creates a migration runner.
func NewRunner(db *gorm.DB, source Source, options ...Option) Runner {
	runner := Runner{
		db:       db,
		source:   source,
		store:    NewStore(db),
		locker:   NewLocker(db),
		log:      zap.NewNop(),
		executor: defaultExecutor(),
	}
	for _, option := range options {
		option(&runner)
	}
	return runner
}

// Status describes current migration state.
type Status struct {
	// Applied contains applied migrations.
	Applied []Record

	// Pending contains pending migrations.
	Pending []Migration

	// Dirty reports whether the database has dirty state.
	Dirty bool
}

// Up applies all pending migrations.
func (runner Runner) Up(ctx context.Context) (Status, error) {
	if err := runner.prepare(ctx); err != nil {
		return Status{}, err
	}
	defer runner.unlock(ctx)

	migrations, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	if err := validateRecords(migrations, records); err != nil {
		return Status{}, err
	}
	pending := pendingMigrations(migrations, records)
	status := Status{Applied: records, Pending: pending}
	for _, migration := range pending {
		if err := runner.apply(ctx, migration); err != nil {
			return status, err
		}
	}
	return runner.Status(ctx)
}

// Validate verifies migration files and applied history without executing SQL.
func (runner Runner) Validate(ctx context.Context) (Status, error) {
	if err := runner.store.Ensure(ctx); err != nil {
		return Status{}, err
	}
	migrations, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	if err := validateRecords(migrations, records); err != nil {
		return Status{}, err
	}
	return Status{Applied: records, Pending: pendingMigrations(migrations, records)}, nil
}

// Status returns current migration state.
func (runner Runner) Status(ctx context.Context) (Status, error) {
	if err := runner.store.Ensure(ctx); err != nil {
		return Status{}, err
	}
	migrations, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	return Status{Applied: records, Pending: pendingMigrations(migrations, records), Dirty: dirty(records)}, nil
}

// Repair clears dirty migration state after manual repair.
func (runner Runner) Repair(ctx context.Context, version int64, checksum string, reason string) error {
	if err := runner.prepare(ctx); err != nil {
		return err
	}
	defer runner.unlock(ctx)
	return runner.store.Repair(ctx, version, checksum, reason)
}

// Down rolls back applied migrations by steps.
func (runner Runner) Down(ctx context.Context, steps int) (Status, error) {
	if steps < 1 {
		return Status{}, fmt.Errorf("steps must be greater than zero")
	}
	if err := runner.prepare(ctx); err != nil {
		return Status{}, err
	}
	defer runner.unlock(ctx)

	migrations, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	if err := validateRecords(migrations, records); err != nil {
		return Status{}, err
	}
	targets := rollbackTargets(migrations, records, steps)
	for _, migration := range targets {
		if err := runner.rollback(ctx, migration); err != nil {
			return Status{}, err
		}
	}
	return runner.Status(ctx)
}

// Reset rolls back all applied migrations.
func (runner Runner) Reset(ctx context.Context) (Status, error) {
	status, err := runner.Status(ctx)
	if err != nil {
		return Status{}, err
	}
	if len(status.Applied) == 0 {
		return status, nil
	}
	return runner.Down(ctx, len(status.Applied))
}

// prepare ensures state and acquires the migration lock.
func (runner Runner) prepare(ctx context.Context) error {
	if err := runner.store.Ensure(ctx); err != nil {
		return err
	}
	return runner.locker.Lock(ctx)
}

// unlock releases the migration lock and logs failures.
func (runner Runner) unlock(ctx context.Context) {
	if err := runner.locker.Unlock(ctx); err != nil {
		runner.log.Error("release migration lock failed", zap.Error(err))
	}
}

// loadState loads migration files and applied records.
func (runner Runner) loadState(ctx context.Context) ([]Migration, []Record, error) {
	migrations, err := Load(runner.source)
	if err != nil {
		return nil, nil, err
	}
	records, err := runner.store.Applied(ctx)
	if err != nil {
		return nil, nil, err
	}
	return migrations, records, nil
}

// apply executes one migration and records its state.
func (runner Runner) apply(ctx context.Context, migration Migration) error {
	started := time.Now().UTC()
	runner.log.Info("applying database migration", zap.Int64("version", migration.Version), zap.String("name", migration.Name), zap.String("checksum", migration.Checksum))
	if err := runner.store.Start(ctx, migration, runner.executor, runner.appVersion); err != nil {
		return err
	}
	err := runner.execute(ctx, migration)
	if err != nil {
		_ = runner.store.Fail(ctx, migration, started, err)
		return err
	}
	if err := runner.store.Succeed(ctx, migration, started); err != nil {
		return err
	}
	runner.log.Info("database migration applied", zap.Int64("version", migration.Version), zap.String("name", migration.Name), zap.Int64("duration_ms", time.Since(started).Milliseconds()))
	return nil
}

// execute runs migration SQL.
func (runner Runner) execute(ctx context.Context, migration Migration) error {
	sql := migrationSQL(runner.db.Dialector.Name(), migration.UpSQL)
	if !migration.Transaction {
		return runner.db.WithContext(ctx).Exec(sql).Error
	}
	return runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Exec(sql).Error
	})
}

// rollback executes a down migration and removes its history row.
func (runner Runner) rollback(ctx context.Context, migration Migration) error {
	sql := migrationSQL(runner.db.Dialector.Name(), migration.DownSQL)
	runner.log.Info("rolling back database migration", zap.Int64("version", migration.Version), zap.String("name", migration.Name))
	err := runner.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Exec(sql).Error
	})
	if err != nil {
		return err
	}
	return runner.store.Delete(ctx, migration.Version)
}

// migrationSQL returns SQL adapted for the active dialect.
func migrationSQL(dialect string, script string) string {
	if dialect != "sqlite" {
		return script
	}
	script = strings.ReplaceAll(script, "timestamptz", "datetime")
	script = strings.ReplaceAll(script, "uuid", "text")
	script = strings.ReplaceAll(script, "jsonb", "text")
	return stripPostgresOnlyLines(script)
}

// stripPostgresOnlyLines removes clauses unsupported by SQLite test databases.
func stripPostgresOnlyLines(script string) string {
	lines := strings.Split(script, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "USING gin") || strings.Contains(line, "to_tsvector") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// validateRecords validates applied records against migration files.
func validateRecords(migrations []Migration, records []Record) error {
	byVersion := map[int64]Migration{}
	for _, migration := range migrations {
		byVersion[migration.Version] = migration
	}
	for _, record := range records {
		if record.Dirty {
			return fmt.Errorf("%w: version %06d", ErrDirty, record.Version)
		}
		migration, ok := byVersion[record.Version]
		if !ok {
			return fmt.Errorf("applied migration %06d has no migration file", record.Version)
		}
		if migration.Checksum != record.Checksum {
			return fmt.Errorf("%w: version %06d", ErrChecksumChanged, record.Version)
		}
	}
	return nil
}

// pendingMigrations returns migrations without applied records.
func pendingMigrations(migrations []Migration, records []Record) []Migration {
	applied := map[int64]struct{}{}
	for _, record := range records {
		if record.Success && !record.Dirty {
			applied[record.Version] = struct{}{}
		}
	}
	var pending []Migration
	for _, migration := range migrations {
		if _, ok := applied[migration.Version]; !ok {
			pending = append(pending, migration)
		}
	}
	return pending
}

// rollbackTargets returns applied migrations in rollback order.
func rollbackTargets(migrations []Migration, records []Record, steps int) []Migration {
	byVersion := map[int64]Migration{}
	for _, migration := range migrations {
		byVersion[migration.Version] = migration
	}
	var targets []Migration
	for index := len(records) - 1; index >= 0 && len(targets) < steps; index-- {
		if migration, ok := byVersion[records[index].Version]; ok {
			targets = append(targets, migration)
		}
	}
	return targets
}

// dirty reports whether any record is dirty.
func dirty(records []Record) bool {
	for _, record := range records {
		if record.Dirty {
			return true
		}
	}
	return false
}

// defaultExecutor returns the default migration executor name.
func defaultExecutor() string {
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}
	return "gamehub"
}
