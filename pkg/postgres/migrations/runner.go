package migrations

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrDirty reports that the database has a dirty migration.
var ErrDirty = errors.New("database migration state is dirty")

// ErrChecksumChanged reports that an applied migration checksum changed.
var ErrChecksumChanged = errors.New("migration checksum changed")

// Runner applies and validates global schema migrations.
type Runner struct {
	db         *gorm.DB    // db stores the db value.
	source     Source      // source stores the source value.
	store      Store       // store stores the store value.
	locker     Locker      // locker stores the locker value.
	log        *zap.Logger // log stores the log value.
	executor   string      // executor stores the executor value.
	appVersion string      // appVersion stores the app version value.
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
