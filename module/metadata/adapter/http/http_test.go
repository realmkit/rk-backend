package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	metadatapostgres "github.com/niflaot/gamehub-go/module/metadata/adapter/postgres"
	"github.com/niflaot/gamehub-go/module/metadata/application"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCreateDefinitionReturnsCreated verifies definition creation over HTTP.
func TestCreateDefinitionReturnsCreated(t *testing.T) {
	app := newTestApp(t)
	body := []byte(`{"owner_type":"user","namespace":"profile","key":"motto","name":"Motto","value_type":"single_line_text","rules":{"max_length":80}}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/metadata/metafield-definitions", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-definition")

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
	if res.Header.Get(headers.ETag) == "" {
		t.Fatalf("%s header = empty", headers.ETag)
	}
}

// TestCreateDefinitionRequiresIdempotencyKey verifies retryable creates require idempotency.
func TestCreateDefinitionRequiresIdempotencyKey(t *testing.T) {
	app := newTestApp(t)
	body := []byte(`{"owner_type":"user","namespace":"profile","key":"motto","name":"Motto","value_type":"single_line_text"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/metadata/metafield-definitions", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusBadRequest)
	}
}

// TestSetValueReturnsCanonicalValue verifies value writes are normalized.
func TestSetValueReturnsCanonicalValue(t *testing.T) {
	app := newTestApp(t)
	createDefinition(t, app)
	ownerID := "4d8decb9-2e4a-4cc7-9d76-5ee74a2dbad8"
	body := []byte(`{"value":"Ready"}`)
	req := httptest.NewRequest(fiber.MethodPut, "/api/v1/metadata/owners/user/"+ownerID+"/metafields/profile/motto", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "set-value")

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	value := payload["value"].(map[string]any)
	if value["value"] != "Ready" {
		t.Fatalf("value.value = %v, want Ready", value["value"])
	}
}

// TestSetValueValidationFailureReturnsProblem verifies domain validation maps to problem JSON.
func TestSetValueValidationFailureReturnsProblem(t *testing.T) {
	app := newTestApp(t)
	createDefinition(t, app)
	ownerID := "4d8decb9-2e4a-4cc7-9d76-5ee74a2dbad8"
	body := []byte("{\"value\":\"line\\nbreak\"}")
	req := httptest.NewRequest(fiber.MethodPut, "/api/v1/metadata/owners/user/"+ownerID+"/metafields/profile/motto", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "set-invalid-value")

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusUnprocessableEntity)
	}
	if res.Header.Get(headers.ContentType) != problem.ContentType {
		t.Fatalf("Content-Type = %q, want %q", res.Header.Get(headers.ContentType), problem.ContentType)
	}
}

// TestDefinitionLifecycleExercisesReadUpdateDelete verifies definition HTTP lifecycle.
func TestDefinitionLifecycleExercisesReadUpdateDelete(t *testing.T) {
	app := newTestApp(t)
	created := createDefinitionPayload(t, app)
	id := created["id"].(string)
	version := uint64(created["version"].(float64))

	listReq := httptest.NewRequest(fiber.MethodGet, "/api/v1/metadata/metafield-definitions?owner_type=user", nil)
	listRes, err := app.Test(listReq, -1)
	if err != nil {
		t.Fatalf("list Test() error = %v", err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != fiber.StatusOK {
		t.Fatalf("list StatusCode = %d, want %d", listRes.StatusCode, fiber.StatusOK)
	}

	getReq := httptest.NewRequest(fiber.MethodGet, "/api/v1/metadata/metafield-definitions/"+id, nil)
	getRes, err := app.Test(getReq, -1)
	if err != nil {
		t.Fatalf("get Test() error = %v", err)
	}
	defer getRes.Body.Close()
	if getRes.StatusCode != fiber.StatusOK {
		t.Fatalf("get StatusCode = %d, want %d", getRes.StatusCode, fiber.StatusOK)
	}

	patchReq := httptest.NewRequest(fiber.MethodPatch, "/api/v1/metadata/metafield-definitions/"+id, bytes.NewReader([]byte(`{"name":"Public Motto"}`)))
	patchReq.Header.Set(headers.ContentType, "application/json")
	patchReq.Header.Set(headers.IfMatch, quoteVersion(version))
	patchRes, err := app.Test(patchReq, -1)
	if err != nil {
		t.Fatalf("patch Test() error = %v", err)
	}
	defer patchRes.Body.Close()
	if patchRes.StatusCode != fiber.StatusOK {
		t.Fatalf("patch StatusCode = %d, want %d", patchRes.StatusCode, fiber.StatusOK)
	}

	deleteReq := httptest.NewRequest(fiber.MethodDelete, "/api/v1/metadata/metafield-definitions/"+id, nil)
	deleteReq.Header.Set(headers.IfMatch, quoteVersion(version+1))
	deleteRes, err := app.Test(deleteReq, -1)
	if err != nil {
		t.Fatalf("delete Test() error = %v", err)
	}
	defer deleteRes.Body.Close()
	if deleteRes.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete StatusCode = %d, want %d", deleteRes.StatusCode, fiber.StatusNoContent)
	}
}

// TestMetaobjectLifecycleExercisesRoutes verifies metaobject definition and entry routes.
func TestMetaobjectLifecycleExercisesRoutes(t *testing.T) {
	app := newTestApp(t)
	body := []byte(`{"type":"profile_card","name":"Profile Card","fields":[{"key":"motto","name":"Motto","value_type":"single_line_text","required":true}]}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/metadata/metaobject-definitions", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-metaobject-definition")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("definition Test() error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("definition StatusCode = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
	definition := decodeObject(t, res)
	definitionID := definition["id"].(string)

	entryBody := []byte(`{"handle":"first_card","display_name":"First Card","fields":{"motto":"Ready"}}`)
	entryReq := httptest.NewRequest(fiber.MethodPost, "/api/v1/metadata/metaobject-definitions/"+definitionID+"/entries", bytes.NewReader(entryBody))
	entryReq.Header.Set(headers.ContentType, "application/json")
	entryReq.Header.Set(headers.IdempotencyKey, "create-metaobject-entry")
	entryRes, err := app.Test(entryReq, -1)
	if err != nil {
		t.Fatalf("entry Test() error = %v", err)
	}
	defer entryRes.Body.Close()
	if entryRes.StatusCode != fiber.StatusCreated {
		t.Fatalf("entry StatusCode = %d, want %d", entryRes.StatusCode, fiber.StatusCreated)
	}
	entry := decodeObject(t, entryRes)
	entryID := entry["id"].(string)
	entryVersion := uint64(entry["version"].(float64))

	listReq := httptest.NewRequest(fiber.MethodGet, "/api/v1/metadata/metaobject-definitions/"+definitionID+"/entries", nil)
	listRes, err := app.Test(listReq, -1)
	if err != nil {
		t.Fatalf("list entries Test() error = %v", err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != fiber.StatusOK {
		t.Fatalf("list entries StatusCode = %d, want %d", listRes.StatusCode, fiber.StatusOK)
	}

	patchReq := httptest.NewRequest(fiber.MethodPatch, "/api/v1/metadata/metaobject-definitions/"+definitionID+"/entries/"+entryID, bytes.NewReader([]byte(`{"display_name":"Updated Card","fields":{"motto":"Still ready"}}`)))
	patchReq.Header.Set(headers.ContentType, "application/json")
	patchReq.Header.Set(headers.IfMatch, quoteVersion(entryVersion))
	patchRes, err := app.Test(patchReq, -1)
	if err != nil {
		t.Fatalf("patch entry Test() error = %v", err)
	}
	defer patchRes.Body.Close()
	if patchRes.StatusCode != fiber.StatusOK {
		t.Fatalf("patch entry StatusCode = %d, want %d", patchRes.StatusCode, fiber.StatusOK)
	}
}

// newTestApp creates a Fiber app with metadata routes.
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := metadatapostgres.Migrate(db); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	store := orm.NewStore(db)
	service := application.NewService(application.Dependencies{
		Definitions:           metadatapostgres.NewMetafieldDefinitionRepository(store),
		Values:                metadatapostgres.NewMetafieldValueRepository(store),
		MetaobjectDefinitions: metadatapostgres.NewMetaobjectDefinitionRepository(store),
		MetaobjectEntries:     metadatapostgres.NewMetaobjectEntryRepository(store),
	})
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(headers.Middleware())
	v1 := app.Group("/api/v1")
	v1.Use(headers.RequireJSON())
	Register(v1, Services{Definitions: service, Values: service, Metaobjects: service})
	return app
}

// createDefinition creates the default test definition.
func createDefinition(t *testing.T, app *fiber.App) {
	t.Helper()
	res := createDefinitionResponse(t, app)
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("create definition StatusCode = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
}

// createDefinitionPayload creates the default definition and returns its payload.
func createDefinitionPayload(t *testing.T, app *fiber.App) map[string]any {
	t.Helper()
	res := createDefinitionResponse(t, app)
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("create definition StatusCode = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
	return decodeObject(t, res)
}

// createDefinitionResponse creates the default definition and returns its response.
func createDefinitionResponse(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()
	body := []byte(`{"owner_type":"user","namespace":"profile","key":"motto","name":"Motto","value_type":"single_line_text"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/metadata/metafield-definitions", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-definition")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	return res
}

// decodeObject decodes response body into an object.
func decodeObject(t *testing.T, res *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// quoteVersion formats a numeric version for If-Match.
func quoteVersion(version uint64) string {
	return `"` + strconv.FormatUint(version, 10) + `"`
}
