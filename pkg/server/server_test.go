package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	metadatahttp "github.com/niflaot/gamehub-go/module/metadata/adapter/http"
	metadatapostgres "github.com/niflaot/gamehub-go/module/metadata/adapter/postgres"
	"github.com/niflaot/gamehub-go/module/metadata/application"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/openapi"
	"github.com/niflaot/gamehub-go/pkg/api/swagger"
	"github.com/niflaot/gamehub-go/pkg/api/versioning"
	"github.com/niflaot/gamehub-go/pkg/logger"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestConfigAddress verifies server addresses are formatted for Listen.
func TestConfigAddress(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 9090}

	if cfg.Address() != "127.0.0.1:9090" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "127.0.0.1:9090")
	}
}

// TestNewServesHealth verifies the server exposes a health endpoint.
func TestNewServesHealth(t *testing.T) {
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

// TestNewServesVersionedHealth verifies v1 routes are registered centrally.
func TestNewServesVersionedHealth(t *testing.T) {
	app := New(nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/api/v1/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}

// TestNewWritesRequestHeaders verifies common response headers are applied.
func TestNewWritesRequestHeaders(t *testing.T) {
	app := New(nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/api/v1/health", nil)

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

// TestNewUsesZapFiberMiddleware verifies Fiber access logs are emitted as JSON.
func TestNewUsesZapFiberMiddleware(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(logger.Config{Level: "info"}, logger.WithOutput(&output))
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	app := New(log, false)
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
	development := New(nil, true)
	production := New(nil, false)

	if development.Config().DisableStartupMessage {
		t.Fatalf("development DisableStartupMessage = true, want false")
	}
	if !production.Config().DisableStartupMessage {
		t.Fatalf("production DisableStartupMessage = false, want true")
	}
}

// TestNewServesSwaggerOnlyInDevelopment verifies Swagger follows the development gate.
func TestNewServesSwaggerOnlyInDevelopment(t *testing.T) {
	development := New(nil, true)
	production := New(nil, false)

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
	app := New(nil, true)

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
	app := New(nil, true, WithMetadata(newMetadataServices(t)))

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
	if route.Path == versioning.V1.Prefix {
		return false
	}
	if route.Path == swagger.DocsPath || route.Path == swagger.OpenAPIPath {
		return false
	}
	return route.Method != "USE"
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
	service := application.NewService(application.Dependencies{
		Definitions:           metadatapostgres.NewMetafieldDefinitionRepository(store),
		Values:                metadatapostgres.NewMetafieldValueRepository(store),
		MetaobjectDefinitions: metadatapostgres.NewMetaobjectDefinitionRepository(store),
		MetaobjectEntries:     metadatapostgres.NewMetaobjectEntryRepository(store),
	})
	return metadatahttp.Services{Definitions: service, Values: service, Metaobjects: service}
}
