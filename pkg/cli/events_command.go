package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	forumsapp "github.com/realmkit/rk-backend/module/forums/application"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/pkg/events/application"
	realmkitredis "github.com/realmkit/rk-backend/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
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
					if _, err := service.DispatchOnce(ctx, "realmkit-cli"); err != nil {
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
			result, err := runEventsReport(
				cmd.Context(),
				activeLogger,
				deps,
				func(ctx context.Context, service application.Service) (application.DispatchResult, error) {
					return service.DispatchOnce(ctx, "realmkit-cli")
				},
			)
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
func runEventsReport(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	action func(context.Context, application.Service) (application.DispatchResult, error),
) (application.DispatchResult, error) {
	var result application.DispatchResult
	err := runEventsAction(ctx, activeLogger, deps, func(ctx context.Context, service application.Service) error {
		var err error
		result, err = action(ctx, service)
		return err
	})
	return result, err
}

// runEventsAction runs one event service action.
func runEventsAction(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	action func(context.Context, application.Service) error,
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
	client, err := deps.openRedis(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	defer func() {
		if err := deps.closeRedis(client); err != nil {
			log.Error("close redis failed", zap.Error(err))
		}
	}()
	return action(ctx, eventsService(db, client, nil))
}

// runForumReport runs a forum operation returning a drift report.
func runForumReport(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	needsRedis bool,
	action func(context.Context, forumsapp.Service) (forumsdomain.CounterDriftReport, error),
) (forumsdomain.CounterDriftReport, error) {
	var report forumsdomain.CounterDriftReport
	err := runForumAction(ctx, activeLogger, deps, needsRedis, func(ctx context.Context, service forumsapp.Service) error {
		var err error
		report, err = action(ctx, service)
		return err
	})
	return report, err
}

// runForumAction runs a forum operational action.
func runForumAction(
	ctx context.Context,
	activeLogger **zap.Logger,
	deps commandDeps,
	needsRedis bool,
	action func(context.Context, forumsapp.Service) error,
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
	client, closeRedis, err := optionalRedis(ctx, deps, cfg.Redis, needsRedis)
	if err != nil {
		return err
	}
	defer closeRedis(log)
	events := eventsService(db, client, nil)
	punishments := punishmentsService(db, client, events)
	return action(ctx, forumsService(db, client, nil, punishments, events))
}

// optionalRedis opens Redis only when a command needs it.
func optionalRedis(
	ctx context.Context,
	deps commandDeps,
	cfg realmkitredis.Config,
	enabled bool,
) (*goredis.Client, func(*zap.Logger), error) {
	if !enabled {
		return nil, func(*zap.Logger) {}, nil
	}
	client, err := deps.openRedis(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, func(log *zap.Logger) {
		if err := deps.closeRedis(client); err != nil {
			log.Error("close redis failed", zap.Error(err))
		}
	}, nil
}

// writeForumReport writes a counter drift report.
func writeForumReport(output io.Writer, report forumsdomain.CounterDriftReport) {
	fmt.Fprintf(output, "mismatches=%d repaired=%s\n", len(report.Mismatches), strconv.FormatBool(report.Repaired))
	for _, mismatch := range report.Mismatches {
		fmt.Fprintf(
			output,
			"drift object_type=%s object_id=%s field=%s expected=%d actual=%d\n",
			mismatch.ObjectType,
			mismatch.ObjectID,
			mismatch.Field,
			mismatch.Expected,
			mismatch.Actual,
		)
	}
}
