package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
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
	realmkitcors "github.com/realmkit/rk-backend/pkg/api/cors"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/idempotency"
	"github.com/realmkit/rk-backend/pkg/api/openapi"
	"github.com/realmkit/rk-backend/pkg/api/ratelimit"
	"github.com/realmkit/rk-backend/pkg/api/swagger"
	"github.com/realmkit/rk-backend/pkg/logger"
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

// TestNewServesAuthConfig verifies public auth config route wiring.
func TestNewServesAuthConfig(t *testing.T) {
	authConfig, userService, userServices := newUserServices(t)
	app := newApp(t, nil, true, WithAuth(authConfig, userService), WithUsers(userServices))
	req := httptest.NewRequest(fiber.MethodGet, "/auth/config", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusOK)
	}
}

// TestNewUsesDefaultIdempotencyStore verifies isolated server construction succeeds.
func TestNewUsesDefaultIdempotencyStore(t *testing.T) {
	app := New(nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}

// TestConfigAddress verifies server addresses are formatted for Listen.
func TestConfigAddress(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 9090}

	if cfg.Address() != "127.0.0.1:9090" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "127.0.0.1:9090")
	}
}

// TestNewConfiguresLargeHeaderBuffer verifies auth cookies do not break docs.
func TestNewConfiguresLargeHeaderBuffer(t *testing.T) {
	app := newApp(t, nil, true)
	req := httptest.NewRequest(fiber.MethodGet, swagger.DocsPath, nil)
	req.Header.Set("Cookie", "realmkit="+strings.Repeat("x", 24*1024))

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if app.Config().ReadBufferSize != DefaultReadBufferSize {
		t.Fatalf("ReadBufferSize = %d, want %d", app.Config().ReadBufferSize, DefaultReadBufferSize)
	}
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusOK)
	}
}

// TestNewServesHealth verifies the server exposes a health endpoint.
func TestNewServesHealth(t *testing.T) {
	app := newApp(t, nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}

// TestNewRejectsGatewayVersionPrefix verifies RealmKit does not own public version prefixes.
func TestNewRejectsGatewayVersionPrefix(t *testing.T) {
	app := newApp(t, nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/api"+"/v1/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNotFound)
	}
}

// TestNewWritesRequestHeaders verifies common response headers are applied.
func TestNewWritesRequestHeaders(t *testing.T) {
	app := newApp(t, nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get(headers.RequestID) == "" {
		t.Fatalf("%s header = empty", headers.RequestID)
	}
	if res.Header.Get(headers.CorrelationID) == "" {
		t.Fatalf("%s header = empty", headers.CorrelationID)
	}
	if res.Header.Get(headers.RateLimitLimit) == "" {
		t.Fatalf("%s header = empty", headers.RateLimitLimit)
	}
}

// TestNewAppliesCORS verifies configured browser origins are allowed.
func TestNewAppliesCORS(t *testing.T) {
	app := newApp(t, nil, false, WithCORS(realmkitcors.Config{Enabled: true, AllowOrigins: "http://localhost:3000"}))
	req := httptest.NewRequest(fiber.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", fiber.MethodGet)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

// TestNewUsesInjectedRateLimitStore verifies server options wire custom rate limit stores.
func TestNewUsesInjectedRateLimitStore(t *testing.T) {
	app := newApp(t, nil, false, WithRateLimitStore(denyingRateLimitStore{}))
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusTooManyRequests)
	}
}

// TestNewUsesZapFiberMiddleware verifies Fiber access logs are emitted as JSON.
func TestNewUsesZapFiberMiddleware(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(logger.Config{Level: "info"}, logger.WithOutput(&output))
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	app := newApp(t, log, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("Unmarshal() error = %v output = %q", err, output.String())
	}
	if entry["level"] != "info" {
		t.Fatalf("level = %v, want %v", entry["level"], "info")
	}
	if entry["status"] != float64(fiber.StatusNoContent) {
		t.Fatalf("status = %v, want %v", entry["status"], fiber.StatusNoContent)
	}
}

// TestNewControlsFiberStartupMessage verifies Fiber's banner is development-only.
func TestNewControlsFiberStartupMessage(t *testing.T) {
	development := newApp(t, nil, true)
	production := newApp(t, nil, false)

	if development.Config().DisableStartupMessage {
		t.Fatalf("development DisableStartupMessage = true, want false")
	}
	if !production.Config().DisableStartupMessage {
		t.Fatalf("production DisableStartupMessage = false, want true")
	}
}

// TestNewServesSwaggerOnlyInDevelopment verifies Swagger follows the development gate.
func TestNewServesSwaggerOnlyInDevelopment(t *testing.T) {
	development := newApp(t, nil, true)
	production := newApp(t, nil, false)

	devRes, err := development.Test(httptest.NewRequest(fiber.MethodGet, swagger.DocsPath, nil), -1)
	if err != nil {
		t.Fatalf("development Test() error = %v", err)
	}
	defer devRes.Body.Close()
	prodRes, err := production.Test(httptest.NewRequest(fiber.MethodGet, swagger.DocsPath, nil), -1)
	if err != nil {
		t.Fatalf("production Test() error = %v", err)
	}
	defer prodRes.Body.Close()

	if devRes.StatusCode != fiber.StatusOK {
		t.Fatalf("development StatusCode = %d, want %d", devRes.StatusCode, fiber.StatusOK)
	}
	if prodRes.StatusCode != fiber.StatusNotFound {
		t.Fatalf("production StatusCode = %d, want %d", prodRes.StatusCode, fiber.StatusNotFound)
	}
}

// TestRegisteredPublicRoutesExistInOpenAPI verifies Fiber routes are documented.
func TestRegisteredPublicRoutesExistInOpenAPI(t *testing.T) {
	app := newApp(t, nil, true)

	for _, route := range app.GetRoutes() {
		if !requiresContract(route) {
			continue
		}

		ok, err := openapi.OperationExists(route.Method, route.Path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !ok {
			t.Fatalf("%s %s missing OpenAPI operation", route.Method, route.Path)
		}
	}
}

// TestRegisteredMetadataRoutesExistInOpenAPI verifies optional metadata routes are documented.
func TestRegisteredMetadataRoutesExistInOpenAPI(t *testing.T) {
	app := newApp(t, nil, true, WithMetadata(newMetadataServices(t)))

	for _, route := range app.GetRoutes() {
		if !requiresContract(route) {
			continue
		}

		ok, err := openapi.OperationExists(route.Method, route.Path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !ok {
			t.Fatalf("%s %s missing OpenAPI operation", route.Method, route.Path)
		}
	}
}

// TestRegisteredGroupsRoutesExistInOpenAPI verifies optional groups routes are documented.
func TestRegisteredGroupsRoutesExistInOpenAPI(t *testing.T) {
	app := newApp(t, nil, true, WithGroups(newGroupsServices(t)))

	for _, route := range app.GetRoutes() {
		if !requiresContract(route) {
			continue
		}

		ok, err := openapi.OperationExists(route.Method, route.Path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !ok {
			t.Fatalf("%s %s missing OpenAPI operation", route.Method, route.Path)
		}
	}
}

// TestRegisteredUserRoutesExistInOpenAPI verifies optional user routes are documented.
func TestRegisteredUserRoutesExistInOpenAPI(t *testing.T) {
	authConfig, userService, userServices := newUserServices(t)
	app := newApp(t, nil, true, WithAuth(authConfig, userService), WithUsers(userServices))

	for _, route := range app.GetRoutes() {
		if !requiresContract(route) {
			continue
		}

		ok, err := openapi.OperationExists(route.Method, route.Path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !ok {
			t.Fatalf("%s %s missing OpenAPI operation", route.Method, route.Path)
		}
	}
}

// requiresContract reports whether route must exist in OpenAPI.
func requiresContract(route fiber.Route) bool {
	if route.Method == fiber.MethodHead {
		return false
	}
	if route.Path == "/" {
		return false
	}
	if route.Path == "/users" {
		return false
	}
	if route.Path == swagger.DocsPath || route.Path == swagger.OpenAPIPath {
		return false
	}
	return route.Method != "USE"
}

// newApp creates a server with Redis-backed idempotency for tests.
func newApp(t *testing.T, log *zap.Logger, development bool, opts ...Option) *fiber.App {
	t.Helper()
	options := []Option{WithIdempotencyStore(newRedisIdempotencyStore(t))}
	options = append(options, opts...)
	return New(log, development, options...)
}

// newRedisIdempotencyStore creates a Redis idempotency store for server tests.
func newRedisIdempotencyStore(t *testing.T) idempotency.RedisStore {
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
func newMetadataServices(t *testing.T) metadatahttp.Services {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
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
func newGroupsServices(t *testing.T) groupshttp.Services {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	store := orm.NewStore(db)
	service := groupsapplication.NewService(
		groupspostgres.NewGroupRepository(store),
		groupspostgres.NewMembershipRepository(store),
		groupspostgres.NewTupleRepository(store),
		groupspostgres.NewPermissionRepository(store),
	)
	return groupshttp.Services{Groups: service, Memberships: service, Tuples: service, Checker: service}
}

// newUserServices creates auth config and user services for server tests.
func newUserServices(t *testing.T) (auth.Config, userapplication.Service, userhttp.Services) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
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
