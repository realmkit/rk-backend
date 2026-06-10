package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/events/application"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newEventsCommand creates the event operations command group.
func newEventsCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "events",
		Short:         "Operate durable events",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newEventsDispatchCommand(activeLogger, deps))
	cmd.AddCommand(newEventsDispatchOnceCommand(activeLogger, deps))
	cmd.AddCommand(newEventsReplayCommand(activeLogger, deps))
	cmd.AddCommand(newEventsCancelCommand(activeLogger, deps))
	return cmd
}

// newEventsDispatchCommand creates the dispatch worker command.
func newEventsDispatchCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:           "dispatch",
		Short:         "Dispatch events until stopped",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEventsAction(cmd.Context(), activeLogger, deps, func(ctx context.Context, service application.Service) error {
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				for {
					if _, err := service.DispatchOnce(ctx, "gamehub-cli"); err != nil {
						return err
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
					}
				}
			})
		},
	}
}

// newEventsDispatchOnceCommand creates the dispatch-once command.
func newEventsDispatchOnceCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:           "dispatch-once",
		Short:         "Dispatch one event batch",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := runEventsReport(cmd.Context(), activeLogger, deps, func(ctx context.Context, service application.Service) (application.DispatchResult, error) {
				return service.DispatchOnce(ctx, "gamehub-cli")
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "claimed=%d processed=%d failed=%d\n", result.Claimed, result.Processed, result.Failed)
			return nil
		},
	}
}

// newEventsReplayCommand creates the replay command.
func newEventsReplayCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:           "replay {event_id}",
		Short:         "Replay one event",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}
			return runEventsAction(cmd.Context(), activeLogger, deps, func(ctx context.Context, service application.Service) error {
				return service.Replay(ctx, id)
			})
		},
	}
}

// newEventsCancelCommand creates the cancel command.
func newEventsCancelCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:           "cancel {event_id}",
		Short:         "Cancel one event",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}
			return runEventsAction(cmd.Context(), activeLogger, deps, func(ctx context.Context, service application.Service) error {
				return service.Cancel(ctx, id)
			})
		},
	}
}

// runEventsReport runs an event command returning a dispatch result.
func runEventsReport(ctx context.Context, activeLogger **zap.Logger, deps commandDeps, action func(context.Context, application.Service) (application.DispatchResult, error)) (application.DispatchResult, error) {
	var result application.DispatchResult
	err := runEventsAction(ctx, activeLogger, deps, func(ctx context.Context, service application.Service) error {
		var err error
		result, err = action(ctx, service)
		return err
	})
	return result, err
}

// runEventsAction runs one event service action.
func runEventsAction(ctx context.Context, activeLogger **zap.Logger, deps commandDeps, action func(context.Context, application.Service) error) error {
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
	return action(ctx, eventsService(db, client))
}
