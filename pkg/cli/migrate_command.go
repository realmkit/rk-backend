package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"

	forumsapp "github.com/realmkit/rk-backend/module/forums/application"
	groupsapp "github.com/realmkit/rk-backend/module/groups/application"
	punishmentsapp "github.com/realmkit/rk-backend/module/punishments/application"
	ticketsdomain "github.com/realmkit/rk-backend/module/tickets/domain"
	cronhttp "github.com/realmkit/rk-backend/pkg/cronjob/adapter/http"
	eventshttp "github.com/realmkit/rk-backend/pkg/events/adapter/http"
	eventsapp "github.com/realmkit/rk-backend/pkg/events/application"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// newMigrateCommand creates the migrate command group.
func newMigrateCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "migrate",
		Short:         "Manage PostgreSQL schema migrations",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newMigrateUpCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateStatusCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateValidateCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateRepairCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateDownCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateResetCommand(activeLogger, deps))
	return cmd
}

// newMigrateUpCommand creates the migrate up command.
func newMigrateUpCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return migrateStatusCommand(
		"up",
		"Apply pending migrations",
		activeLogger,
		deps,
		func(ctx context.Context, runner migrations.Runner) (migrations.Status, error) {
			return runner.Up(ctx)
		},
	)
}

// newMigrateStatusCommand creates the migrate status command.
func newMigrateStatusCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return migrateStatusCommand(
		"status",
		"Show migration status",
		activeLogger,
		deps,
		func(ctx context.Context, runner migrations.Runner) (migrations.Status, error) {
			return runner.Status(ctx)
		},
	)
}

// newMigrateValidateCommand creates the migrate validate command.
func newMigrateValidateCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return migrateStatusCommand(
		"validate",
		"Validate migration files and history",
		activeLogger,
		deps,
		func(ctx context.Context, runner migrations.Runner) (migrations.Status, error) {
			return runner.Validate(ctx)
		},
	)
}

// migrateStatusCommand creates a migration command that writes status.
func migrateStatusCommand(
	use string,
	short string,
	activeLogger **zap.Logger,
	deps commandDeps,
	action func(context.Context, migrations.Runner) (migrations.Status, error),
) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			status, err := runMigration(cmd.Context(), activeLogger, deps, action)
			if err != nil {
				return err
			}
			writeStatus(cmd.OutOrStdout(), status)
			return nil
		},
	}
}

// newMigrateRepairCommand creates the migrate repair command.
func newMigrateRepairCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	var version int64
	var checksum string
	var reason string
	cmd := &cobra.Command{
		Use:           "repair",
		Short:         "Clear dirty migration state after manual repair",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version == 0 || checksum == "" || reason == "" {
				return fmt.Errorf("version, checksum, and reason are required")
			}
			return runMigrationRepair(cmd.Context(), activeLogger, deps, version, checksum, reason)
		},
	}
	cmd.Flags().Int64Var(&version, "version", 0, "migration version to repair")
	cmd.Flags().StringVar(&checksum, "checksum", "", "expected migration checksum")
	cmd.Flags().StringVar(&reason, "reason", "", "manual repair reason")
	return cmd
}

// newMigrateDownCommand creates the migrate down command.
func newMigrateDownCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	var steps int
	var confirmed bool
	cmd := migrateStatusCommand(
		"down",
		"Roll back applied migrations",
		activeLogger,
		deps,
		func(ctx context.Context, runner migrations.Runner) (migrations.Status, error) {
			if !confirmed {
				return migrations.Status{}, fmt.Errorf("down requires --i-understand-this-can-destroy-data")
			}
			return runner.Down(ctx, steps)
		},
	)
	cmd.Flags().IntVar(&steps, "steps", 1, "number of migrations to roll back")
	cmd.Flags().BoolVar(&confirmed, "i-understand-this-can-destroy-data", false, "confirm destructive rollback")
	return cmd
}

// newMigrateResetCommand creates the migrate reset command.
func newMigrateResetCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	var confirmed bool
	cmd := migrateStatusCommand(
		"reset",
		"Roll back all applied migrations",
		activeLogger,
		deps,
		func(ctx context.Context, runner migrations.Runner) (migrations.Status, error) {
			if !confirmed {
				return migrations.Status{}, fmt.Errorf("reset requires --i-understand-this-can-destroy-data")
			}
			return runner.Reset(ctx)
		},
	)
	cmd.Flags().BoolVar(&confirmed, "i-understand-this-can-destroy-data", false, "confirm destructive reset")
	return cmd
}

// runMigration runs a migration command that returns status.
func runMigration(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	action func(context.Context, migrations.Runner) (migrations.Status, error),
) (migrations.Status, error) {
	cfg, log, err := runtime(ctx, activeLogger, deps)
	if err != nil {
		return migrations.Status{}, err
	}
	db, err := deps.openPostgres(ctx, cfg.Postgres)
	if err != nil {
		return migrations.Status{}, err
	}
	defer closeDatabase(log, deps.closePostgres, db)
	return action(ctx, deps.newRunner(db, log))
}

// runMigrationRepair runs the migration repair command.
func runMigrationRepair(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	version int64,
	checksum string,
	reason string,
) error {
	cfg, log, err := runtime(ctx, activeLogger, deps)
	if err != nil {
		return err
	}
	db, err := deps.openPostgres(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer closeDatabase(log, deps.closePostgres, db)
	return deps.newRunner(db, log).Repair(ctx, version, checksum, reason)
}

// writeStatus writes migration status to output.
func writeStatus(output io.Writer, status migrations.Status) {
	fmt.Fprintf(output, "applied=%d pending=%d dirty=%s\n", len(status.Applied), len(status.Pending), strconv.FormatBool(status.Dirty))
	for _, migration := range status.Pending {
		fmt.Fprintf(output, "pending %06d %s\n", migration.Version, migration.Name)
	}
}

// infrastructureOptions creates server options for shared infrastructure.
func infrastructureOptions(_ context.Context, db *gorm.DB, events eventsapp.Service, hub *eventshttp.Hub, groups groupsapp.Service, forums forumsapp.Service, punishments punishmentsapp.Service, tickets ticketOperations) ([]server.Option, error) {
	cron := cronService(db, events, forums, punishments, tickets)
	return []server.Option{
		server.WithEvents(eventshttp.Services{Events: events, Hub: hub, Checker: groups}),
		server.WithCron(cronhttp.Services{Cron: cron, Checker: groups}),
	}, nil
}

// ticketOperations is the ticket surface used by cron wiring.
type ticketOperations interface {
	DetectSLABreaches(context.Context) (int64, error)
	CloseStaleTickets(context.Context) (int64, error)
	VerifyStats(context.Context) (ticketsdomain.DriftReport, error)
	RebuildStats(context.Context) (ticketsdomain.DriftReport, error)
}

// closeDatabase closes a database and logs failures.
func closeDatabase(log *zap.Logger, closePostgres func(*gorm.DB) error, db *gorm.DB) {
	if err := closePostgres(db); err != nil {
		log.Error("close postgres failed", zap.Error(err))
	}
}
