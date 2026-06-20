// Package seedcmd owns RealmKit seed CLI commands.
package seedcmd

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/google/uuid"
	themepostgres "github.com/realmkit/rk-backend/module/themes/adapter/postgres"
	themeapplication "github.com/realmkit/rk-backend/module/themes/application"
	"github.com/realmkit/rk-backend/pkg/config"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres"
	"github.com/realmkit/rk-backend/pkg/postgres/seeding"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Dependencies contains seed command runtime dependencies.
type Dependencies struct {
	LoadConfig    func() (config.Config, error)
	NewLogger     func(logger.Config) (*zap.Logger, error)
	OpenPostgres  func(context.Context, postgres.Config) (*gorm.DB, error)
	ClosePostgres func(*gorm.DB) error
}

// New creates the seed command group.
func New(
	activeLogger **zap.Logger,
	loadConfig func() (config.Config, error),
	newLogger func(logger.Config) (*zap.Logger, error),
	openPostgres func(context.Context, postgres.Config) (*gorm.DB, error),
	closePostgres func(*gorm.DB) error,
) *cobra.Command {
	deps := Dependencies{
		LoadConfig:    loadConfig,
		NewLogger:     newLogger,
		OpenPostgres:  openPostgres,
		ClosePostgres: closePostgres,
	}
	cmd := &cobra.Command{
		Use:           "seed",
		Short:         "Manage global data seeds",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newUpCommand(activeLogger, deps))
	cmd.AddCommand(newStatusCommand(activeLogger, deps))
	cmd.AddCommand(newValidateCommand(activeLogger, deps))
	cmd.AddCommand(newDryRunCommand(activeLogger, deps))
	cmd.AddCommand(newRepairCommand(activeLogger, deps))
	cmd.AddCommand(newGrantAdminCommand(activeLogger, deps))
	return cmd
}

// SeedThemeSigningKeys stores environment-backed theme signing keys.
func SeedThemeSigningKeys(ctx context.Context, db *gorm.DB, cfg config.Config) error {
	repository := themepostgres.NewSigningKeyRepository(orm.NewStore(db))
	return themeapplication.SeedSigningKeys(ctx, repository, cfg.Themes)
}

// newUpCommand creates the seed up command.
func newUpCommand(activeLogger **zap.Logger, deps Dependencies) *cobra.Command {
	return statusCommand("up", "Apply pending data seeds", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Up(ctx)
	})
}

// newStatusCommand creates the seed status command.
func newStatusCommand(activeLogger **zap.Logger, deps Dependencies) *cobra.Command {
	return statusCommand("status", "Show data seed status", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Status(ctx)
	})
}

// newDryRunCommand creates the seed dry-run command.
func newDryRunCommand(activeLogger **zap.Logger, deps Dependencies) *cobra.Command {
	return statusCommand("dry-run", "Show pending data seeds without mutation", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Validate(ctx)
	})
}

// newValidateCommand creates the seed validate command.
func newValidateCommand(activeLogger **zap.Logger, deps Dependencies) *cobra.Command {
	return statusCommand("validate", "Validate data seed files and history", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Validate(ctx)
	})
}

// statusCommand creates a seed command that writes status.
func statusCommand(
	use string,
	short string,
	activeLogger **zap.Logger,
	deps Dependencies,
	action func(context.Context, seeding.Runner) (seeding.Status, error),
) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			status, err := run(cmd.Context(), activeLogger, deps, action)
			if err != nil {
				return err
			}
			writeStatus(cmd.OutOrStdout(), status)
			return nil
		},
	}
}

// newRepairCommand creates the seed repair command.
func newRepairCommand(activeLogger **zap.Logger, deps Dependencies) *cobra.Command {
	var version int64
	var checksum string
	var reason string
	cmd := &cobra.Command{
		Use:           "repair",
		Short:         "Clear dirty seed state after manual repair",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version == 0 || checksum == "" || reason == "" {
				return fmt.Errorf("version, checksum, and reason are required")
			}
			return runRepair(cmd.Context(), activeLogger, deps, version, checksum, reason)
		},
	}
	cmd.Flags().Int64Var(&version, "version", 0, "seed version to repair")
	cmd.Flags().StringVar(&checksum, "checksum", "", "expected seed checksum")
	cmd.Flags().StringVar(&reason, "reason", "", "manual repair reason")
	return cmd
}

// newGrantAdminCommand creates the admin grant command.
func newGrantAdminCommand(activeLogger **zap.Logger, deps Dependencies) *cobra.Command {
	var userID string
	cmd := &cobra.Command{
		Use:           "grant-admin",
		Short:         "Grant the seeded administrator group to a local user",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			parsed, err := uuid.Parse(userID)
			if err != nil {
				return fmt.Errorf("user-id must be a UUID: %w", err)
			}
			grant, err := runGrantAdmin(cmd.Context(), activeLogger, deps, parsed)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "admin_group=%s user=%s created=%s\n", grant.GroupID, grant.UserID, strconv.FormatBool(grant.Created))
			return nil
		},
	}
	cmd.Flags().StringVar(&userID, "user-id", "", "local user UUID to grant")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

// run executes a seed action.
func run(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps Dependencies,
	action func(context.Context, seeding.Runner) (seeding.Status, error),
) (seeding.Status, error) {
	cfg, log, err := runtime(activeLogger, deps)
	if err != nil {
		return seeding.Status{}, err
	}
	db, err := deps.OpenPostgres(ctx, cfg.Postgres)
	if err != nil {
		return seeding.Status{}, err
	}
	defer closeDatabase(log, deps.ClosePostgres, db)
	runner := seeding.NewRunner(db, seeding.DefaultSource(), seeding.WithLogger(log), seeding.WithExecutor("realmkit-cli"))
	return action(ctx, runner)
}

// runRepair runs the seed repair command.
func runRepair(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps Dependencies,
	version int64,
	checksum string,
	reason string,
) error {
	_, err := run(ctx, activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return seeding.Status{}, runner.Repair(ctx, version, checksum, reason)
	})
	return err
}

// runGrantAdmin runs the admin grant command.
func runGrantAdmin(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps Dependencies,
	userID uuid.UUID,
) (seeding.AdminGrant, error) {
	var grant seeding.AdminGrant
	_, err := run(ctx, activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		var grantErr error
		grant, grantErr = runner.GrantAdmin(ctx, userID)
		return seeding.Status{}, grantErr
	})
	return grant, err
}

// runtime loads configuration and creates the active logger.
func runtime(activeLogger **zap.Logger, deps Dependencies) (config.Config, *zap.Logger, error) {
	cfg, err := deps.LoadConfig()
	if err != nil {
		return config.Config{}, nil, err
	}
	log, err := deps.NewLogger(cfg.Logging)
	if err != nil {
		return config.Config{}, nil, err
	}
	*activeLogger = log
	return cfg, log, nil
}

// closeDatabase closes the database and logs failures.
func closeDatabase(log *zap.Logger, close func(*gorm.DB) error, db *gorm.DB) {
	if err := close(db); err != nil {
		log.Error("close postgres failed", zap.Error(err))
	}
}

// writeStatus writes seed status to output.
func writeStatus(output io.Writer, status seeding.Status) {
	fmt.Fprintf(output, "applied=%d pending=%d dirty=%s\n", len(status.Applied), len(status.Pending), strconv.FormatBool(status.Dirty))
	for _, seed := range status.Pending {
		fmt.Fprintf(output, "pending %06d %s\n", seed.Version, seed.Name)
	}
}
