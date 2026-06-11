package cli

import (
	"context"
	"fmt"
	"time"

	cronapp "github.com/niflaot/gamehub-go/pkg/cronjob/application"
	crondefaults "github.com/niflaot/gamehub-go/pkg/cronjob/defaults"
	cronDomain "github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newCronCommand creates the cron command group.
func newCronCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{Use: "cron", Short: "Operate cron jobs", SilenceUsage: true, SilenceErrors: true}
	cmd.AddCommand(newCronRunCommand(activeLogger, deps))
	cmd.AddCommand(newCronRunOnceCommand(activeLogger, deps))
	cmd.AddCommand(newCronTriggerCommand(activeLogger, deps))
	cmd.AddCommand(newCronListCommand(activeLogger, deps))
	cmd.AddCommand(newCronRepairLocksCommand(activeLogger, deps))
	return cmd
}

// newCronRunCommand creates the cron worker command.
func newCronRunCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{Use: "run", Short: "Run due cron jobs until stopped", SilenceUsage: true, SilenceErrors: true, RunE: func(cmd *cobra.Command, _ []string) error {
		return runCronAction(cmd.Context(), activeLogger, deps, func(ctx context.Context, service cronapp.Service) error {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for {
				_, _ = service.RunOnce(ctx)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
				}
			}
		})
	}}
}

// newCronRunOnceCommand creates the run-once command.
func newCronRunOnceCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{Use: "run-once", Short: "Run one due cron job", SilenceUsage: true, SilenceErrors: true, RunE: func(cmd *cobra.Command, _ []string) error {
		return runCronSummary(cmd.Context(), activeLogger, deps, cmd, func(ctx context.Context, service cronapp.Service) (cronapp.RunSummary, error) {
			return service.RunOnce(ctx)
		})
	}}
}

// newCronTriggerCommand creates the trigger command.
func newCronTriggerCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{Use: "trigger {job_key}", Short: "Run one cron job now", Args: cobra.ExactArgs(1), SilenceUsage: true, SilenceErrors: true, RunE: func(cmd *cobra.Command, args []string) error {
		return runCronSummary(cmd.Context(), activeLogger, deps, cmd, func(ctx context.Context, service cronapp.Service) (cronapp.RunSummary, error) {
			return service.Trigger(ctx, args[0])
		})
	}}
}

// newCronListCommand creates the list command.
func newCronListCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List cron jobs", SilenceUsage: true, SilenceErrors: true, RunE: func(cmd *cobra.Command, _ []string) error {
		return runCronAction(cmd.Context(), activeLogger, deps, func(ctx context.Context, service cronapp.Service) error {
			result, err := service.ListDefinitions(ctx, pagination.Page{Limit: 100})
			if err != nil {
				return err
			}
			for _, definition := range result.Items {
				fmt.Fprintf(cmd.OutOrStdout(), "job key=%s enabled=%t next_run_at=%v last_status=%s\n", definition.Key, definition.Enabled, definition.NextRunAt, definition.LastStatus)
			}
			return nil
		})
	}}
}

// newCronRepairLocksCommand creates the repair-locks command.
func newCronRepairLocksCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{Use: "repair-locks", Short: "Repair stale cron locks", SilenceUsage: true, SilenceErrors: true, RunE: func(cmd *cobra.Command, _ []string) error {
		return runCronAction(cmd.Context(), activeLogger, deps, func(ctx context.Context, service cronapp.Service) error {
			count, err := service.RepairLocks(ctx)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "repaired=%d\n", count)
			return nil
		})
	}}
}

// runCronSummary runs a cron command and writes a run summary.
func runCronSummary(ctx context.Context, activeLogger **zap.Logger, deps commandDeps, cmd *cobra.Command, action func(context.Context, cronapp.Service) (cronapp.RunSummary, error)) error {
	return runCronAction(ctx, activeLogger, deps, func(ctx context.Context, service cronapp.Service) error {
		result, err := action(ctx, service)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "run_id=%s job_key=%s processed=%d failed=%t\n", result.RunID, result.JobKey, result.ProcessedCount, result.Failed)
		return nil
	})
}

// runCronAction runs one cron service action.
func runCronAction(ctx context.Context, activeLogger **zap.Logger, deps commandDeps, action func(context.Context, cronapp.Service) error) error {
	cfg, log, err := runtime(ctx, activeLogger, deps)
	if err != nil {
		return err
	}
	db, err := deps.openPostgres(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer closeDatabase(log, deps.closePostgres, db)
	client, err := deps.openRedis(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	defer func() {
		if err := deps.closeRedis(client); err != nil {
			log.Error("close redis failed", zap.Error(err))
		}
	}()
	events := eventsService(db, client, nil)
	punishments := punishmentsService(db, client, events)
	forums := forumsService(db, client, nil, punishments, events)
	tickets := ticketsService(db, client, nil, punishments, events)
	service := cronService(db, events, forums, punishments, tickets)
	if err := service.EnsureDefinitions(ctx, cronDefinitions(time.Now().UTC())); err != nil {
		return err
	}
	return action(ctx, service)
}

// cronDefinitions composes startup-owned cron defaults.
func cronDefinitions(now time.Time) []cronDomain.Definition {
	definitions := crondefaults.EventDefinitions(now)
	definitions = append(definitions, crondefaults.ForumDefinitions(now)...)
	definitions = append(definitions, crondefaults.PunishmentDefinitions(now)...)
	definitions = append(definitions, crondefaults.TicketDefinitions(now)...)
	definitions = append(definitions, crondefaults.MaintenanceDefinitions(now)...)
	return definitions
}
