package seeding

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrDirty reports that the database has a dirty seed.
var ErrDirty = errors.New("database seed state is dirty")

// ErrChecksumChanged reports that an applied seed checksum changed.
var ErrChecksumChanged = errors.New("seed checksum changed")

// ErrUserNotFound reports that an admin grant targets an unknown user.
var ErrUserNotFound = errors.New("user not found for admin seed grant")

// Runner applies and validates global data seeds.
type Runner struct {
	db         *gorm.DB    // db stores the db value.
	source     Source      // source stores the source value.
	store      Store       // store stores the store value.
	log        *zap.Logger // log stores the log value.
	executor   string      // executor stores the executor value.
	appVersion string      // appVersion stores the app version value.
}

// Option configures a Runner.
type Option func(*Runner)

// WithLogger configures seed logging.
func WithLogger(log *zap.Logger) Option {
	return func(runner *Runner) {
		runner.log = log
	}
}

// WithExecutor configures the seed executor name.
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

// NewRunner creates a seed runner.
func NewRunner(db *gorm.DB, source Source, options ...Option) Runner {
	runner := Runner{
		db:       db,
		source:   source,
		store:    NewStore(db),
		log:      zap.NewNop(),
		executor: "realmkit",
	}
	for _, option := range options {
		option(&runner)
	}
	return runner
}

// Status describes current seed state.
type Status struct {
	// Applied contains applied seeds.
	Applied []Record

	// Pending contains pending seeds.
	Pending []Seed

	// Dirty reports whether the database has dirty state.
	Dirty bool
}

// AdminGrant describes an administrator membership grant.
type AdminGrant struct {
	// UserID is the granted user.
	UserID uuid.UUID

	// GroupID is the administrator group.
	GroupID uuid.UUID

	// Created reports whether a membership row was created or restored.
	Created bool
}

// Up applies all pending seeds.
func (runner Runner) Up(ctx context.Context) (Status, error) {
	if err := runner.prepare(ctx); err != nil {
		return Status{}, err
	}
	defer runner.unlock(ctx)

	seeds, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	if err := validateRecords(seeds, records); err != nil {
		return Status{}, err
	}
	pending := pendingSeeds(seeds, records)
	status := Status{Applied: records, Pending: pending}
	for _, seed := range pending {
		if err := runner.apply(ctx, seed); err != nil {
			return status, err
		}
	}
	return runner.Status(ctx)
}

// Validate verifies seed files and applied history without executing SQL.
func (runner Runner) Validate(ctx context.Context) (Status, error) {
	if err := runner.store.Ensure(ctx); err != nil {
		return Status{}, err
	}
	seeds, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	if err := validateRecords(seeds, records); err != nil {
		return Status{}, err
	}
	return Status{Applied: records, Pending: pendingSeeds(seeds, records), Dirty: dirty(records)}, nil
}

// Status returns current seed state.
func (runner Runner) Status(ctx context.Context) (Status, error) {
	if err := runner.store.Ensure(ctx); err != nil {
		return Status{}, err
	}
	seeds, records, err := runner.loadState(ctx)
	if err != nil {
		return Status{}, err
	}
	return Status{Applied: records, Pending: pendingSeeds(seeds, records), Dirty: dirty(records)}, nil
}

// Repair clears dirty seed state after manual repair.
func (runner Runner) Repair(ctx context.Context, version int64, checksum string, reason string) error {
	if err := runner.prepare(ctx); err != nil {
		return err
	}
	defer runner.unlock(ctx)
	return runner.store.Repair(ctx, version, checksum, reason)
}

// GrantAdmin grants the seeded administrator group to one local user.
func (runner Runner) GrantAdmin(ctx context.Context, userID uuid.UUID) (AdminGrant, error) {
	if userID == uuid.Nil {
		return AdminGrant{}, fmt.Errorf("%w: nil user id", ErrUserNotFound)
	}
	if err := runner.prepare(ctx); err != nil {
		return AdminGrant{}, err
	}
	defer runner.unlock(ctx)
	return runner.grantAdmin(ctx, userID)
}

// prepare ensures state and acquires the seed lock.
func (runner Runner) prepare(ctx context.Context) error {
	if err := runner.store.Ensure(ctx); err != nil {
		return err
	}
	return runner.lock(ctx)
}

// lock acquires the seed advisory lock.
func (runner Runner) lock(ctx context.Context) error {
	if runner.db.Dialector.Name() != "postgres" {
		return nil
	}
	return runner.db.WithContext(ctx).Exec("SELECT pg_advisory_lock(hashtext('realmkit_data_seeds'))").Error
}

// unlock releases the seed advisory lock and logs failures.
func (runner Runner) unlock(ctx context.Context) {
	if runner.db.Dialector.Name() != "postgres" {
		return
	}
	if err := runner.db.WithContext(ctx).Exec("SELECT pg_advisory_unlock(hashtext('realmkit_data_seeds'))").Error; err != nil {
		runner.log.Error("release seed lock failed", zap.Error(err))
	}
}

// loadState loads seed files and applied records.
func (runner Runner) loadState(ctx context.Context) ([]Seed, []Record, error) {
	seeds, err := Load(runner.source)
	if err != nil {
		return nil, nil, err
	}
	records, err := runner.store.Applied(ctx)
	if err != nil {
		return nil, nil, err
	}
	return seeds, records, nil
}
