package cli

import (
	"context"
	"time"

	assetshttp "github.com/niflaot/gamehub-go/module/assets/adapter/http"
	assetsapp "github.com/niflaot/gamehub-go/module/assets/application"
	assetsport "github.com/niflaot/gamehub-go/module/assets/port"
	forumsassets "github.com/niflaot/gamehub-go/module/forums/adapter/assets"
	forumshttp "github.com/niflaot/gamehub-go/module/forums/adapter/http"
	forumspostgres "github.com/niflaot/gamehub-go/module/forums/adapter/postgres"
	forumsredis "github.com/niflaot/gamehub-go/module/forums/adapter/redis"
	forumsapp "github.com/niflaot/gamehub-go/module/forums/application"
	forumsport "github.com/niflaot/gamehub-go/module/forums/port"
	groupshttp "github.com/niflaot/gamehub-go/module/groups/adapter/http"
	groupspostgres "github.com/niflaot/gamehub-go/module/groups/adapter/postgres"
	groupsapp "github.com/niflaot/gamehub-go/module/groups/application"
	metadatahttp "github.com/niflaot/gamehub-go/module/metadata/adapter/http"
	metadatapostgres "github.com/niflaot/gamehub-go/module/metadata/adapter/postgres"
	metadataapp "github.com/niflaot/gamehub-go/module/metadata/application"
	userhttp "github.com/niflaot/gamehub-go/module/user/adapter/http"
	userpostgres "github.com/niflaot/gamehub-go/module/user/adapter/postgres"
	userapp "github.com/niflaot/gamehub-go/module/user/application"
	"github.com/niflaot/gamehub-go/pkg/config"
	cronhttp "github.com/niflaot/gamehub-go/pkg/cronjob/adapter/http"
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
	"github.com/niflaot/gamehub-go/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// infrastructureOptions creates server options for shared infrastructure.
func infrastructureOptions(_ context.Context, db *gorm.DB, events eventsapp.Service, forums forumsapp.Service) ([]server.Option, error) {
	cron := cronService(db, events, forums)
	return []server.Option{
		server.WithEvents(eventshttp.Services{Events: events, Hub: eventshttp.NewHub()}),
		server.WithCron(cronhttp.Services{Cron: cron}),
	}, nil
}

// assetshttpServices creates HTTP services for assets.
func assetshttpServices(assetService assetsapp.Service) assetshttp.Services {
	return assetshttp.Services{Assets: assetService}
}

// forumsService creates forums application service.
func forumsService(
	db *gorm.DB,
	client *goredis.Client,
	assetService assetsport.Service,
	events eventsapp.Service,
) forumsapp.Service {
	store := orm.NewStore(db)
	var readCache forumsport.ReadCache
	if client != nil {
		readCache = forumsredis.NewTreeCache(client)
	}
	return forumsapp.NewService(forumsapp.Dependencies{
		Categories:   forumspostgres.NewCategoryRepository(store),
		Forums:       forumspostgres.NewForumRepository(store),
		Threads:      forumspostgres.NewThreadRepository(store),
		Posts:        forumspostgres.NewPostRepository(store),
		Interactions: forumspostgres.NewInteractionRepository(store),
		Operations:   forumspostgres.NewOperationsRepository(store),
		Assets:       forumsassets.NewResolver(assetService),
		Authorizer:   forumspostgres.NewVisibilityAuthorizer(store),
		Cache:        readCache,
		Transactions: transaction.New(db),
		Events:       events,
	})
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

// groupsService creates groups application service.
func groupsService(db *gorm.DB, events eventsapp.Service) groupsapp.Service {
	store := orm.NewStore(db)
	return groupsapp.NewService(
		groupspostgres.NewGroupRepository(store),
		groupspostgres.NewMembershipRepository(store),
		groupspostgres.NewTupleRepository(store),
		groupspostgres.NewPermissionRepository(store),
	).WithEvents(events)
}

// groupshttpServices creates HTTP services for groups.
func groupshttpServices(groupService groupsapp.Service) groupshttp.Services {
	return groupshttp.Services{
		Groups:      groupService,
		Memberships: groupService,
		Tuples:      groupService,
		Checker:     groupService,
	}
}

// metadataService creates metadata application service.
func metadataService(db *gorm.DB, events eventsapp.Service) metadataapp.Service {
	store := orm.NewStore(db)
	return metadataapp.NewService(metadataapp.Dependencies{
		Definitions:           metadatapostgres.NewMetafieldDefinitionRepository(store),
		Values:                metadatapostgres.NewMetafieldValueRepository(store),
		MetaobjectDefinitions: metadatapostgres.NewMetaobjectDefinitionRepository(store),
		MetaobjectEntries:     metadatapostgres.NewMetaobjectEntryRepository(store),
		Events:                events,
	})
}

// metadatahttpServices creates HTTP services for metadata.
func metadatahttpServices(service metadataapp.Service) metadatahttp.Services {
	return metadatahttp.Services{
		Definitions: service,
		Values:      service,
		Metaobjects: service,
	}
}

// usersService creates users application service.
func usersService(db *gorm.DB, cfg config.Config, events eventsapp.Service) userapp.Service {
	store := orm.NewStore(db)
	return userapp.NewService(userapp.Dependencies{
		Users:        userpostgres.NewUserRepository(store),
		Links:        userpostgres.NewIdentityLinkRepository(store),
		Claims:       userpostgres.NewClaimCacheRepository(store),
		Transactions: transaction.New(db),
		Provider:     cfg.Auth.Provider,
		Events:       events,
	})
}

// usershttpServices creates HTTP services for users.
func usershttpServices(userService userapp.Service, groupService groupsapp.Service) userhttp.Services {
	return userhttp.Services{Users: userService, Groups: groupService}
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
		Events:       events,
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
