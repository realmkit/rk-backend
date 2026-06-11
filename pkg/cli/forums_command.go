package cli

import (
	"context"
	"fmt"

	assetshttp "github.com/niflaot/gamehub-go/module/assets/adapter/http"
	assetsport "github.com/niflaot/gamehub-go/module/assets/port"
	forumsassets "github.com/niflaot/gamehub-go/module/forums/adapter/assets"
	forumshttp "github.com/niflaot/gamehub-go/module/forums/adapter/http"
	forumsapp "github.com/niflaot/gamehub-go/module/forums/application"
	forumsdomain "github.com/niflaot/gamehub-go/module/forums/domain"
	punishmentshttp "github.com/niflaot/gamehub-go/module/punishments/adapter/http"
	punishmentsport "github.com/niflaot/gamehub-go/module/punishments/port"
	ticketshttp "github.com/niflaot/gamehub-go/module/tickets/adapter/http"
	ticketspostgres "github.com/niflaot/gamehub-go/module/tickets/adapter/postgres"
	ticketpunishments "github.com/niflaot/gamehub-go/module/tickets/adapter/punishments"
	ticketsredis "github.com/niflaot/gamehub-go/module/tickets/adapter/redis"
	ticketsapp "github.com/niflaot/gamehub-go/module/tickets/application"
	ticketsport "github.com/niflaot/gamehub-go/module/tickets/port"
	eventsapp "github.com/niflaot/gamehub-go/pkg/events/application"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// newForumsCommand creates the forums operational command group.
func newForumsCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "forums",
		Short:         "Operate forum caches and counters",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newForumsStatsCommand(activeLogger, deps))
	cmd.AddCommand(newForumsLikesCommand(activeLogger, deps))
	cmd.AddCommand(newForumsCacheCommand(activeLogger, deps))
	cmd.AddCommand(newForumsViewsCommand(activeLogger, deps))
	return cmd
}

// newForumsStatsCommand creates the forum stats command group.
func newForumsStatsCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{Use: "stats", Short: "Verify or rebuild forum stats", SilenceUsage: true, SilenceErrors: true}
	cmd.AddCommand(
		newForumReportCommand(
			"verify",
			"Verify forum stats counters",
			activeLogger,
			deps,
			false,
			"forum stats drift detected",
			func(ctx context.Context, service forumsapp.Service) (forumsdomain.CounterDriftReport, error) {
				return service.VerifyStats(ctx)
			},
		),
	)
	cmd.AddCommand(
		newForumReportCommand(
			"rebuild",
			"Rebuild forum stats counters",
			activeLogger,
			deps,
			false,
			"",
			func(ctx context.Context, service forumsapp.Service) (forumsdomain.CounterDriftReport, error) {
				return service.RebuildStats(ctx)
			},
		),
	)
	return cmd
}

// newForumsLikesCommand creates the forum likes command group.
func newForumsLikesCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{Use: "likes", Short: "Verify or rebuild forum like counters", SilenceUsage: true, SilenceErrors: true}
	cmd.AddCommand(
		newForumReportCommand(
			"verify",
			"Verify forum like counters",
			activeLogger,
			deps,
			false,
			"forum like drift detected",
			func(ctx context.Context, service forumsapp.Service) (forumsdomain.CounterDriftReport, error) {
				return service.VerifyLikes(ctx)
			},
		),
	)
	cmd.AddCommand(
		newForumReportCommand(
			"rebuild",
			"Rebuild forum like counters",
			activeLogger,
			deps,
			false,
			"",
			func(ctx context.Context, service forumsapp.Service) (forumsdomain.CounterDriftReport, error) {
				return service.RebuildLikes(ctx)
			},
		),
	)
	return cmd
}

// newForumReportCommand creates one forum report command.
func newForumReportCommand(
	use string,
	short string,
	activeLogger **zap.Logger,
	deps commandDeps,
	needsRedis bool,
	driftError string,
	action func(context.Context, forumsapp.Service) (forumsdomain.CounterDriftReport, error),
) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := runForumReport(cmd.Context(), activeLogger, deps, needsRedis, action)
			if err != nil {
				return err
			}
			writeForumReport(cmd.OutOrStdout(), report)
			if driftError != "" && len(report.Mismatches) > 0 {
				return fmt.Errorf("%s", driftError)
			}
			return nil
		},
	}
}

// newForumsCacheCommand creates the forum cache command group.
func newForumsCacheCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{Use: "cache", Short: "Operate forum read caches", SilenceUsage: true, SilenceErrors: true}
	cmd.AddCommand(
		&cobra.Command{
			Use:           "clear",
			Short:         "Clear forum read caches",
			SilenceUsage:  true,
			SilenceErrors: true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				err := runForumAction(cmd.Context(), activeLogger, deps, true, func(ctx context.Context, service forumsapp.Service) error {
					return service.ClearReadCache(ctx)
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), "forum caches cleared")
				return nil
			},
		},
	)
	return cmd
}

// newForumsViewsCommand creates the forum view-counter command group.
func newForumsViewsCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{Use: "views", Short: "Operate forum view counters", SilenceUsage: true, SilenceErrors: true}
	cmd.AddCommand(
		&cobra.Command{
			Use:           "flush",
			Short:         "Flush buffered forum thread views",
			SilenceUsage:  true,
			SilenceErrors: true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				var flushed int64
				err := runForumAction(cmd.Context(), activeLogger, deps, true, func(ctx context.Context, service forumsapp.Service) error {
					count, err := service.FlushThreadViews(ctx)
					flushed = count
					return err
				})
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "flushed_thread_views=%d\n", flushed)
				return nil
			},
		},
	)
	return cmd
}

// assetshttpServices creates HTTP services for assets.
func assetshttpServices(assetService assetsport.Service) assetshttp.Services {
	return assetshttp.Services{Assets: assetService}
}

// forumshttpServices creates HTTP services for forums.
func forumshttpServices(forumService forumsapp.Service) forumshttp.Services {
	return forumshttp.Services{
		Structure:   forumService,
		Content:     forumService,
		Interaction: forumService,
		Operations:  forumService,
		Admin:       forumService,
	}
}

// punishmentshttpServices creates HTTP services for punishments.
func punishmentshttpServices(service punishmentsport.Service) punishmentshttp.Services {
	return punishmentshttp.Services{Punishments: service}
}

// ticketsService creates tickets application service.
func ticketsService(
	db *gorm.DB,
	client *goredis.Client,
	assets assetsport.Service,
	punishments punishmentsport.Service,
	events eventsapp.Service,
) ticketsapp.Service {
	store := orm.NewStore(db)
	var cache ticketsport.Cache
	if client != nil {
		cache = ticketsredis.NewCache(client)
	}
	return ticketsapp.NewService(ticketsapp.Dependencies{
		Definitions:  ticketspostgres.NewDefinitionRepository(store),
		Tickets:      ticketspostgres.NewTicketRepository(store),
		Punishments:  ticketpunishments.NewResolver(punishments),
		Assets:       forumsassets.NewResolver(assets),
		Cache:        cache,
		Transactions: transaction.New(db),
		Events:       events,
	})
}

// ticketshttpServices creates HTTP services for tickets.
func ticketshttpServices(service ticketsapp.Service) ticketshttp.Services {
	return ticketshttp.Services{
		Definitions:  service,
		Tickets:      service,
		Conversation: service,
		Operations:   service,
	}
}
