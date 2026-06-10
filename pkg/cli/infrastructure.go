package cli

import (
	"context"
	"time"

	forumsapp "github.com/niflaot/gamehub-go/module/forums/application"
	"github.com/niflaot/gamehub-go/pkg/cronjob/adapter/http"
	cronpostgres "github.com/niflaot/gamehub-go/pkg/cronjob/adapter/postgres"
	cronapp "github.com/niflaot/gamehub-go/pkg/cronjob/application"
	cronDomain "github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	cronport "github.com/niflaot/gamehub-go/pkg/cronjob/port"
	eventshttp "github.com/niflaot/gamehub-go/pkg/events/adapter/http"
	eventspostgres "github.com/niflaot/gamehub-go/pkg/events/adapter/postgres"
	eventsredis "github.com/niflaot/gamehub-go/pkg/events/adapter/redis"
	eventsapp "github.com/niflaot/gamehub-go/pkg/events/application"
	eventsport "github.com/niflaot/gamehub-go/pkg/events/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/server"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// infrastructureOptions creates server options for shared infrastructure.
func infrastructureOptions(_ context.Context, db *gorm.DB, client *goredis.Client, forums forumsapp.Service) ([]server.Option, error) {
	events := eventsService(db, client)
	cron := cronService(db, events, forums)
	return []server.Option{
		server.WithEvents(eventshttp.Services{Events: events, Hub: eventshttp.NewHub()}),
		server.WithCron(http.Services{Cron: cron}),
	}, nil
}

// eventsService creates the events application service.
func eventsService(db *gorm.DB, client *goredis.Client) eventsapp.Service {
	var broker eventsport.Broker
	if client != nil {
		broker = eventsredis.NewBroker(client)
	}
	return eventsapp.NewService(eventsapp.Dependencies{
		Repository: eventspostgres.NewRepository(orm.NewStore(db)),
		Broker:     broker,
	})
}

// cronService creates the cron application service.
func cronService(db *gorm.DB, events eventsapp.Service, forums forumsapp.Service) cronapp.Service {
	handlers := map[string]cronport.Handler{
		cronDomain.JobEventsDispatchPending: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			result, err := events.DispatchOnce(ctx, "cron-events-dispatch")
			return cronDomain.Result{ProcessedCount: int64(result.Processed)}, err
		}),
		cronDomain.JobForumsFlushThreadViews: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			count, err := forums.FlushThreadViews(ctx)
			return cronDomain.Result{ProcessedCount: count, ChangedCount: count}, err
		}),
		cronDomain.JobForumsVerifyStats: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			report, err := forums.VerifyStats(ctx)
			return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches))}, err
		}),
		cronDomain.JobForumsVerifyLikes: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			report, err := forums.VerifyLikes(ctx)
			return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches))}, err
		}),
		cronDomain.JobPunishmentsExpireActive:    noopHandler(),
		cronDomain.JobAssetsExpireUploadIntents:  noopHandler(),
		cronDomain.JobUsersCleanupIdentityClaims: noopHandler(),
	}
	return cronapp.NewService(cronapp.Dependencies{
		Repository:   cronpostgres.NewRepository(orm.NewStore(db)),
		WorkerID:     "gamehub-http",
		LockDuration: 5 * time.Minute,
	}, handlers)
}

// noopHandler returns a handler for planned modules that are not implemented yet.
func noopHandler() cronport.Handler {
	return cronport.HandlerFunc(func(context.Context, cronport.RunContext) (cronDomain.Result, error) {
		return cronDomain.Result{SkippedCount: 1}, nil
	})
}
