package server

import (
	"context"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	groupshttp "github.com/realmkit/rk-backend/module/groups/adapter/http"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapplication "github.com/realmkit/rk-backend/module/groups/application"
	metadatahttp "github.com/realmkit/rk-backend/module/metadata/adapter/http"
	metadatapostgres "github.com/realmkit/rk-backend/module/metadata/adapter/postgres"
	metadataapplication "github.com/realmkit/rk-backend/module/metadata/application"
	userhttp "github.com/realmkit/rk-backend/module/user/adapter/http"
	userpostgres "github.com/realmkit/rk-backend/module/user/adapter/postgres"
	userapplication "github.com/realmkit/rk-backend/module/user/application"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/idempotency"
	"github.com/realmkit/rk-backend/pkg/api/ratelimit"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"github.com/realmkit/rk-backend/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// denyingRateLimitStore rejects every request.
type denyingRateLimitStore struct{}

// Allow returns a denied rate limit decision.
func (store denyingRateLimitStore) Allow(context.Context, string, ratelimit.Policy) (ratelimit.Decision, error) {
	return ratelimit.Decision{
		Allowed: false,
		Limit:   1,
		ResetAt: time.Now().Add(time.Minute),
	}, nil
}

// newApp creates a server with Redis-backed idempotency for tests.
func newApp(t testingT, log *zap.Logger, development bool, opts ...Option) *fiber.App {
	t.Helper()
	options := []Option{WithIdempotencyStore(newRedisIdempotencyStore(t))}
	options = append(options, opts...)
	return New(log, development, options...)
}

// newRedisIdempotencyStore creates a Redis idempotency store for server tests.
func newRedisIdempotencyStore(t testingT) idempotency.RedisStore {
	t.Helper()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
	return idempotency.NewRedisStore(client, idempotency.WithRedisScope("server-test"))
}

// newMetadataServices creates metadata services for server tests.
func newMetadataServices(t testingT) metadatahttp.Services {
	t.Helper()
	db := newServerTestDB(t)
	store := orm.NewStore(db)
	service := metadataapplication.NewService(metadataapplication.Dependencies{
		Definitions:           metadatapostgres.NewMetafieldDefinitionRepository(store),
		Values:                metadatapostgres.NewMetafieldValueRepository(store),
		MetaobjectDefinitions: metadatapostgres.NewMetaobjectDefinitionRepository(store),
		MetaobjectEntries:     metadatapostgres.NewMetaobjectEntryRepository(store),
	})
	return metadatahttp.Services{Definitions: service, Values: service, Metaobjects: service}
}

// newGroupsServices creates groups services for server tests.
func newGroupsServices(t testingT) groupshttp.Services {
	t.Helper()
	db := newServerTestDB(t)
	store := orm.NewStore(db)
	service := groupsapplication.NewService(
		groupspostgres.NewGroupRepository(store),
		groupspostgres.NewMembershipRepository(store),
		groupspostgres.NewPermissionRepository(store),
	)
	return groupshttp.Services{Groups: service, Memberships: service, Grants: service, Checker: service}
}

// newUserServices creates auth config and user services for server tests.
func newUserServices(t testingT) (auth.Config, userapplication.Service, userhttp.Services) {
	t.Helper()
	db := newServerTestDB(t)
	store := orm.NewStore(db)
	service := userapplication.NewService(userapplication.Dependencies{
		Users:        userpostgres.NewUserRepository(store),
		Links:        userpostgres.NewIdentityLinkRepository(store),
		Claims:       userpostgres.NewClaimCacheRepository(store),
		Transactions: transaction.New(db),
		Provider:     "generic_oidc",
	})
	config := auth.Config{
		Provider:          "generic_oidc",
		IssuerURL:         "http://localhost:3001",
		Audience:          "realmkit-api",
		ClientID:          "realmkit-frontend",
		Scopes:            "openid profile email",
		DevelopmentBypass: true,
	}
	return config, service, userhttp.Services{Users: service}
}

// newServerTestDB creates a migrated in-memory database.
func newServerTestDB(t testingT) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return db
}

// testingT is the subset of testing.TB used by server helpers.
type testingT interface {
	Helper()
	Cleanup(func())
	Fatalf(string, ...any)
	Logf(string, ...any)
}
