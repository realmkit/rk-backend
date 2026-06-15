package cli

import (
	"context"
	"time"

	assetsport "github.com/realmkit/rk-backend/module/assets/port"
	forumsassets "github.com/realmkit/rk-backend/module/forums/adapter/assets"
	forumspostgres "github.com/realmkit/rk-backend/module/forums/adapter/postgres"
	forumsredis "github.com/realmkit/rk-backend/module/forums/adapter/redis"
	forumsapp "github.com/realmkit/rk-backend/module/forums/application"
	forumsport "github.com/realmkit/rk-backend/module/forums/port"
	groupshttp "github.com/realmkit/rk-backend/module/groups/adapter/http"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapp "github.com/realmkit/rk-backend/module/groups/application"
	metadatahttp "github.com/realmkit/rk-backend/module/metadata/adapter/http"
	metadatapostgres "github.com/realmkit/rk-backend/module/metadata/adapter/postgres"
	metadataapp "github.com/realmkit/rk-backend/module/metadata/application"
	punishmentspostgres "github.com/realmkit/rk-backend/module/punishments/adapter/postgres"
	punishmentsredis "github.com/realmkit/rk-backend/module/punishments/adapter/redis"
	punishmentsapp "github.com/realmkit/rk-backend/module/punishments/application"
	punishmentsport "github.com/realmkit/rk-backend/module/punishments/port"
	userhttp "github.com/realmkit/rk-backend/module/user/adapter/http"
	userpostgres "github.com/realmkit/rk-backend/module/user/adapter/postgres"
	userapp "github.com/realmkit/rk-backend/module/user/application"
	"github.com/realmkit/rk-backend/pkg/config"
	cronpostgres "github.com/realmkit/rk-backend/pkg/cronjob/adapter/postgres"
	cronapp "github.com/realmkit/rk-backend/pkg/cronjob/application"
	cronDomain "github.com/realmkit/rk-backend/pkg/cronjob/domain"
	cronport "github.com/realmkit/rk-backend/pkg/cronjob/port"
	eventspostgres "github.com/realmkit/rk-backend/pkg/events/adapter/postgres"
	eventsredis "github.com/realmkit/rk-backend/pkg/events/adapter/redis"
	eventsapp "github.com/realmkit/rk-backend/pkg/events/application"
	eventsport "github.com/realmkit/rk-backend/pkg/events/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// forumsService creates forums application service.
func forumsService(
	db *gorm.DB,
	client *goredis.Client,
	assetService assetsport.Service,
	restrictions forumsport.RestrictionChecker,
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
		Restrictions: restrictions,
		Cache:        readCache,
		Transactions: transaction.New(db),
		Events:       events,
	})
}

// punishmentsService creates punishment application service.
func punishmentsService(db *gorm.DB, client *goredis.Client, events eventsapp.Service) punishmentsapp.Service {
	store := orm.NewStore(db)
	definitions := punishmentspostgres.NewDefinitionRepository(store)
	cases := punishmentspostgres.NewCaseRepository(store)
	var restrictionCache punishmentsport.RestrictionCache
	if client != nil {
		restrictionCache = punishmentsredis.NewCache(client)
	}
	return punishmentsapp.NewService(punishmentsapp.Dependencies{
		Definitions:  definitions,
		Cases:        cases,
		Cache:        restrictionCache,
		Transactions: transaction.New(db),
		Events:       events,
	})
}

// groupsService creates groups application service.
func groupsService(db *gorm.DB, events eventsapp.Service) groupsapp.Service {
	store := orm.NewStore(db)
	return groupsapp.NewService(
		groupspostgres.NewGroupRepository(store),
		groupspostgres.NewMembershipRepository(store),
		groupspostgres.NewPermissionRepository(store),
	).WithEvents(events)
}

// groupshttpServices creates HTTP services for groups.
func groupshttpServices(groupService groupsapp.Service, userService userapp.Service) groupshttp.Services {
	return groupshttp.Services{
		Groups:      groupService,
		Memberships: groupService,
		Grants:      groupService,
		Checker:     groupService,
		Users:       userService,
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
func metadatahttpServices(service metadataapp.Service, checker groupsapp.Service) metadatahttp.Services {
	return metadatahttp.Services{
		Definitions: service,
		Values:      service,
		Metaobjects: service,
		Checker:     checker,
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
	return userhttp.Services{Users: userService, Groups: groupService, Checker: groupService}
}

// eventsService creates the events application service.
func eventsService(db *gorm.DB, client *goredis.Client, broker eventsport.Broker) eventsapp.Service {
	if broker == nil && client != nil {
		broker = eventsredis.NewBroker(client)
	}
	return eventsapp.NewService(eventsapp.Dependencies{
		Repository: eventspostgres.NewRepository(orm.NewStore(db)),
		Broker:     broker,
	})
}

// cronService creates the cron application service.
func cronService(
	db *gorm.DB,
	events eventsapp.Service,
	forums forumsapp.Service,
	punishments punishmentsapp.Service,
	tickets ticketOperations,
) cronapp.Service {
	handlers := map[string]cronport.Handler{
		cronDomain.JobEventsDispatchPending: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				result, err := events.DispatchOnce(ctx, "cron-events-dispatch")
				return cronDomain.Result{ProcessedCount: int64(result.Processed)}, err
			},
		),
		cronDomain.JobForumsFlushThreadViews: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				count, err := forums.FlushThreadViews(ctx)
				return cronDomain.Result{ProcessedCount: count, ChangedCount: count}, err
			},
		),
		cronDomain.JobForumsVerifyStats: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			report, err := forums.VerifyStats(ctx)
			return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches))}, err
		}),
		cronDomain.JobForumsVerifyLikes: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			report, err := forums.VerifyLikes(ctx)
			return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches))}, err
		}),
		cronDomain.JobPunishmentsExpireActive: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				count, err := punishments.ExpirePunishments(ctx)
				return cronDomain.Result{ProcessedCount: count, ChangedCount: count}, err
			},
		),
		cronDomain.JobPunishmentsVerifyRestrictions: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				report, err := punishments.VerifyRestrictions(ctx)
				return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches))}, err
			},
		),
		cronDomain.JobPunishmentsRebuildRestrictions: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				report, err := punishments.RebuildRestrictions(ctx)
				return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches)), ChangedCount: int64(len(report.Mismatches))}, err
			},
		),
		cronDomain.JobTicketsDetectSLABreaches: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				count, err := tickets.DetectSLABreaches(ctx)
				return cronDomain.Result{ProcessedCount: count}, err
			},
		),
		cronDomain.JobTicketsCloseStale: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			count, err := tickets.CloseStaleTickets(ctx)
			return cronDomain.Result{ProcessedCount: count, ChangedCount: count}, err
		}),
		cronDomain.JobTicketsVerifyStats: cronport.HandlerFunc(func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
			report, err := tickets.VerifyStats(ctx)
			return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches))}, err
		}),
		cronDomain.JobTicketsRebuildStats: cronport.HandlerFunc(
			func(ctx context.Context, _ cronport.RunContext) (cronDomain.Result, error) {
				report, err := tickets.RebuildStats(ctx)
				return cronDomain.Result{ProcessedCount: int64(len(report.Mismatches)), ChangedCount: int64(len(report.Mismatches))}, err
			},
		),
		cronDomain.JobAssetsExpireUploadIntents:  noopHandler(),
		cronDomain.JobUsersCleanupIdentityClaims: noopHandler(),
	}
	return cronapp.NewService(cronapp.Dependencies{
		Repository:   cronpostgres.NewRepository(orm.NewStore(db)),
		Events:       events,
		WorkerID:     "realmkit-http",
		LockDuration: 5 * time.Minute,
	}, handlers)
}

// noopHandler returns a handler for planned modules that are not implemented yet.
func noopHandler() cronport.Handler {
	return cronport.HandlerFunc(func(context.Context, cronport.RunContext) (cronDomain.Result, error) {
		return cronDomain.Result{SkippedCount: 1}, nil
	})
}
