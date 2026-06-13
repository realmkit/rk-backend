package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/postgres/seeding"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newSeedCommand creates the seed command group.
func newSeedCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "seed",
		Short:         "Manage global data seeds",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newSeedUpCommand(activeLogger, deps))
	cmd.AddCommand(newSeedStatusCommand(activeLogger, deps))
	cmd.AddCommand(newSeedValidateCommand(activeLogger, deps))
	cmd.AddCommand(newSeedDryRunCommand(activeLogger, deps))
	cmd.AddCommand(newSeedRepairCommand(activeLogger, deps))
	cmd.AddCommand(newSeedGrantAdminCommand(activeLogger, deps))
	return cmd
}

// newSeedUpCommand creates the seed up command.
func newSeedUpCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return seedStatusCommand("up", "Apply pending data seeds", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Up(ctx)
	})
}

// newSeedStatusCommand creates the seed status command.
func newSeedStatusCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return seedStatusCommand("status", "Show data seed status", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Status(ctx)
	})
}

// newSeedDryRunCommand creates the seed dry-run command.
func newSeedDryRunCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return seedStatusCommand("dry-run", "Show pending data seeds without mutation", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Validate(ctx)
	})
}

// newSeedValidateCommand creates the seed validate command.
func newSeedValidateCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return seedStatusCommand("validate", "Validate data seed files and history", activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return runner.Validate(ctx)
	})
}

// seedStatusCommand creates a seed command that writes status.
func seedStatusCommand(
	use string,
	short string,
	activeLogger **zap.Logger,
	deps commandDeps,
	action func(context.Context, seeding.Runner) (seeding.Status, error),
) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			status, err := runSeed(cmd.Context(), activeLogger, deps, action)
			if err != nil {
				return err
			}
			writeSeedStatus(cmd.OutOrStdout(), status)
			return nil
		},
	}
}

// newSeedRepairCommand creates the seed repair command.
func newSeedRepairCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
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
			return runSeedRepair(cmd.Context(), activeLogger, deps, version, checksum, reason)
		},
	}
	cmd.Flags().Int64Var(&version, "version", 0, "seed version to repair")
	cmd.Flags().StringVar(&checksum, "checksum", "", "expected seed checksum")
	cmd.Flags().StringVar(&reason, "reason", "", "manual repair reason")
	return cmd
}

// newSeedGrantAdminCommand creates the admin grant command.
func newSeedGrantAdminCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
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
			grant, err := runSeedGrantAdmin(cmd.Context(), activeLogger, deps, parsed)
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

// runSeed runs a seed command that returns status.
func runSeed(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	action func(context.Context, seeding.Runner) (seeding.Status, error),
) (seeding.Status, error) {
	cfg, log, err := runtime(ctx, activeLogger, deps)
	if err != nil {
		return seeding.Status{}, err
	}
	db, err := deps.openPostgres(ctx, cfg.Postgres)
	if err != nil {
		return seeding.Status{}, err
	}
	defer closeDatabase(log, deps.closePostgres, db)
	return action(ctx, deps.newSeedRunner(db, log))
}

// runSeedRepair runs the seed repair command.
func runSeedRepair(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	version int64,
	checksum string,
	reason string,
) error {
	_, err := runSeed(ctx, activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		return seeding.Status{}, runner.Repair(ctx, version, checksum, reason)
	})
	return err
}

// runSeedGrantAdmin runs the admin grant command.
func runSeedGrantAdmin(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	userID uuid.UUID,
) (seeding.AdminGrant, error) {
	var grant seeding.AdminGrant
	_, err := runSeed(ctx, activeLogger, deps, func(ctx context.Context, runner seeding.Runner) (seeding.Status, error) {
		var grantErr error
		grant, grantErr = runner.GrantAdmin(ctx, userID)
		return seeding.Status{}, grantErr
	})
	return grant, err
}

// writeSeedStatus writes seed status to output.
func writeSeedStatus(output io.Writer, status seeding.Status) {
	fmt.Fprintf(output, "applied=%d pending=%d dirty=%s\n", len(status.Applied), len(status.Pending), strconv.FormatBool(status.Dirty))
	for _, seed := range status.Pending {
		fmt.Fprintf(output, "pending %06d %s\n", seed.Version, seed.Name)
	}
}
