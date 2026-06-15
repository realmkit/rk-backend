package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	metadatapostgres "github.com/realmkit/rk-backend/module/metadata/adapter/postgres"
	"github.com/realmkit/rk-backend/module/metadata/application"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCreateDefinitionReturnsCreated verifies definition creation over HTTP.
func TestCreateDefinitionReturnsCreated(t *testing.T) {
	app := newTestApp(t)
	body := []byte(
		`{"owner_type":"user","namespace":"profile","key":"motto","name":"Motto","value_type":"single_line_text","rules":{"max_length":80}}`,
	)
	req := httptest.NewRequest(fiber.MethodPost, "/metadata/metafield-definitions", bytes.NewReader(body))
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
	req := httptest.NewRequest(fiber.MethodPost, "/metadata/metafield-definitions", bytes.NewReader(body))
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

// TestOptionalExpectedVersionParsesHeaders verifies optional version parsing.
func TestOptionalExpectedVersionParsesHeaders(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Get("/version", func(ctx *fiber.Ctx) error {
		version, err := optionalExpectedVersion(ctx)
		if err != nil {
			return err
		}
		if version == nil {
			return ctx.SendStatus(fiber.StatusNoContent)
		}
		return ctx.SendString(strconv.FormatUint(*version, 10))
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/version", nil), -1)
	if err != nil {
		t.Fatalf("Test() no header error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}

	req := httptest.NewRequest(fiber.MethodGet, "/version", nil)
	req.Header.Set(headers.IfMatch, "bad")
	res, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() bad header error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusBadRequest)
	}

	req = httptest.NewRequest(fiber.MethodGet, "/version", nil)
	req.Header.Set(headers.IfMatch, `"7"`)
	res, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() version header error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusOK)
	}
}

// TestHandleErrorMapsValidationAndReferenced verifies HTTP problem mapping.
func TestHandleErrorMapsValidationAndReferenced(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Get("/validation", func(ctx *fiber.Ctx) error {
		return handleError(ctx, domain.ValidationError{Violations: []domain.Violation{{Field: "name", Message: "is required"}}})
	})
	app.Get("/referenced", func(ctx *fiber.Ctx) error {
		return handleError(ctx, port.ErrReferenced)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/validation", nil), -1)
	if err != nil {
		t.Fatalf("Test() validation error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusUnprocessableEntity)
	}

	res, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/referenced", nil), -1)
	if err != nil {
		t.Fatalf("Test() referenced error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusConflict {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusConflict)
	}
}

// TestSetValueReturnsCanonicalValue verifies value writes are normalized.
func TestSetValueReturnsCanonicalValue(t *testing.T) {
	app := newTestApp(t)
	createDefinition(t, app)
	ownerID := "4d8decb9-2e4a-4cc7-9d76-5ee74a2dbad8"
	body := []byte(`{"value":"Ready"}`)
	req := httptest.NewRequest(fiber.MethodPut, "/metadata/owners/user/"+ownerID+"/metafields/profile/motto", bytes.NewReader(body))
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

// TestValueLifecycleExercisesReadListAndDelete verifies owner value HTTP lifecycle.
func TestValueLifecycleExercisesReadListAndDelete(t *testing.T) {
	app := newTestApp(t)
	createDefinition(t, app)
	ownerID := "4d8decb9-2e4a-4cc7-9d76-5ee74a2dbad8"
	value := setOwnerValue(t, app, ownerID, "Ready")
	version := uint64(value["version"].(float64))

	listReq := httptest.NewRequest(
		fiber.MethodGet,
		"/metadata/owners/user/"+ownerID+"/metafields?namespace=profile&include_empty=false",
		nil,
	)
	listRes, err := app.Test(listReq, -1)
	if err != nil {
		t.Fatalf("list values Test() error = %v", err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != fiber.StatusOK {
		t.Fatalf("list values StatusCode = %d, want %d", listRes.StatusCode, fiber.StatusOK)
	}

	getReq := httptest.NewRequest(fiber.MethodGet, "/metadata/owners/user/"+ownerID+"/metafields/profile/motto", nil)
	getRes, err := app.Test(getReq, -1)
	if err != nil {
		t.Fatalf("get value Test() error = %v", err)
	}
	defer getRes.Body.Close()
	if getRes.StatusCode != fiber.StatusOK {
		t.Fatalf("get value StatusCode = %d, want %d", getRes.StatusCode, fiber.StatusOK)
	}

	deleteReq := httptest.NewRequest(fiber.MethodDelete, "/metadata/owners/user/"+ownerID+"/metafields/profile/motto", nil)
	deleteReq.Header.Set(headers.IfMatch, quoteVersion(version))
	deleteRes, err := app.Test(deleteReq, -1)
	if err != nil {
		t.Fatalf("delete value Test() error = %v", err)
	}
	defer deleteRes.Body.Close()
	if deleteRes.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete value StatusCode = %d, want %d", deleteRes.StatusCode, fiber.StatusNoContent)
	}
}

// TestSetValueValidationFailureReturnsProblem verifies domain validation maps to problem JSON.
func TestSetValueValidationFailureReturnsProblem(t *testing.T) {
	app := newTestApp(t)
	createDefinition(t, app)
	ownerID := "4d8decb9-2e4a-4cc7-9d76-5ee74a2dbad8"
	body := []byte("{\"value\":\"line\\nbreak\"}")
	req := httptest.NewRequest(fiber.MethodPut, "/metadata/owners/user/"+ownerID+"/metafields/profile/motto", bytes.NewReader(body))
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

	listReq := httptest.NewRequest(fiber.MethodGet, "/metadata/metafield-definitions?owner_type=user", nil)
	listRes, err := app.Test(listReq, -1)
	if err != nil {
		t.Fatalf("list Test() error = %v", err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != fiber.StatusOK {
		t.Fatalf("list StatusCode = %d, want %d", listRes.StatusCode, fiber.StatusOK)
	}

	getReq := httptest.NewRequest(fiber.MethodGet, "/metadata/metafield-definitions/"+id, nil)
	getRes, err := app.Test(getReq, -1)
	if err != nil {
		t.Fatalf("get Test() error = %v", err)
	}
	defer getRes.Body.Close()
	if getRes.StatusCode != fiber.StatusOK {
		t.Fatalf("get StatusCode = %d, want %d", getRes.StatusCode, fiber.StatusOK)
	}

	patchReq := httptest.NewRequest(
		fiber.MethodPatch,
		"/metadata/metafield-definitions/"+id,
		bytes.NewReader([]byte(`{"name":"Public Motto"}`)),
	)
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

	deleteReq := httptest.NewRequest(fiber.MethodDelete, "/metadata/metafield-definitions/"+id, nil)
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
	body := []byte(
		`{"type":"profile_card","name":"Profile Card","fields":[{"key":"motto","name":"Motto","value_type":"single_line_text","required":true}]}`,
	)
	req := httptest.NewRequest(fiber.MethodPost, "/metadata/metaobject-definitions", bytes.NewReader(body))
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
	definitionVersion := uint64(definition["version"].(float64))

	definitionListReq := httptest.NewRequest(fiber.MethodGet, "/metadata/metaobject-definitions?type=profile_card&active=true", nil)
	definitionListRes, err := app.Test(definitionListReq, -1)
	if err != nil {
		t.Fatalf("list definitions Test() error = %v", err)
	}
	defer definitionListRes.Body.Close()
	if definitionListRes.StatusCode != fiber.StatusOK {
		t.Fatalf("list definitions StatusCode = %d, want %d", definitionListRes.StatusCode, fiber.StatusOK)
	}

	definitionGetReq := httptest.NewRequest(fiber.MethodGet, "/metadata/metaobject-definitions/"+definitionID, nil)
	definitionGetRes, err := app.Test(definitionGetReq, -1)
	if err != nil {
		t.Fatalf("get definition Test() error = %v", err)
	}
	defer definitionGetRes.Body.Close()
	if definitionGetRes.StatusCode != fiber.StatusOK {
		t.Fatalf("get definition StatusCode = %d, want %d", definitionGetRes.StatusCode, fiber.StatusOK)
	}

	definitionPatchBody := []byte(`{"name":"Profile Card Updated","description":"Shown on profiles","active":true}`)
	definitionPatchReq := httptest.NewRequest(
		fiber.MethodPatch,
		"/metadata/metaobject-definitions/"+definitionID,
		bytes.NewReader(definitionPatchBody),
	)
	definitionPatchReq.Header.Set(headers.ContentType, "application/json")
	definitionPatchReq.Header.Set(headers.IfMatch, quoteVersion(definitionVersion))
	definitionPatchRes, err := app.Test(definitionPatchReq, -1)
	if err != nil {
		t.Fatalf("patch definition Test() error = %v", err)
	}
	defer definitionPatchRes.Body.Close()
	if definitionPatchRes.StatusCode != fiber.StatusOK {
		t.Fatalf("patch definition StatusCode = %d, want %d", definitionPatchRes.StatusCode, fiber.StatusOK)
	}
	patchedDefinition := decodeObject(t, definitionPatchRes)
	definitionVersion = uint64(patchedDefinition["version"].(float64))

	entryBody := []byte(`{"handle":"first_card","display_name":"First Card","fields":{"motto":"Ready"}}`)
	entryReq := httptest.NewRequest(
		fiber.MethodPost,
		"/metadata/metaobject-definitions/"+definitionID+"/entries",
		bytes.NewReader(entryBody),
	)
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

	listReq := httptest.NewRequest(fiber.MethodGet, "/metadata/metaobject-definitions/"+definitionID+"/entries", nil)
	listRes, err := app.Test(listReq, -1)
	if err != nil {
		t.Fatalf("list entries Test() error = %v", err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != fiber.StatusOK {
		t.Fatalf("list entries StatusCode = %d, want %d", listRes.StatusCode, fiber.StatusOK)
	}

	getEntryReq := httptest.NewRequest(fiber.MethodGet, "/metadata/metaobject-definitions/"+definitionID+"/entries/"+entryID, nil)
	getEntryRes, err := app.Test(getEntryReq, -1)
	if err != nil {
		t.Fatalf("get entry Test() error = %v", err)
	}
	defer getEntryRes.Body.Close()
	if getEntryRes.StatusCode != fiber.StatusOK {
		t.Fatalf("get entry StatusCode = %d, want %d", getEntryRes.StatusCode, fiber.StatusOK)
	}

	patchReq := httptest.NewRequest(
		fiber.MethodPatch,
		"/metadata/metaobject-definitions/"+definitionID+"/entries/"+entryID,
		bytes.NewReader([]byte(`{"display_name":"Updated Card","fields":{"motto":"Still ready"}}`)),
	)
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
	patchedEntry := decodeObject(t, patchRes)
	entryVersion = uint64(patchedEntry["version"].(float64))

	deleteEntryReq := httptest.NewRequest(fiber.MethodDelete, "/metadata/metaobject-definitions/"+definitionID+"/entries/"+entryID, nil)
	deleteEntryReq.Header.Set(headers.IfMatch, quoteVersion(entryVersion))
	deleteEntryRes, err := app.Test(deleteEntryReq, -1)
	if err != nil {
		t.Fatalf("delete entry Test() error = %v", err)
	}
	defer deleteEntryRes.Body.Close()
	if deleteEntryRes.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete entry StatusCode = %d, want %d", deleteEntryRes.StatusCode, fiber.StatusNoContent)
	}

	deleteDefinitionReq := httptest.NewRequest(fiber.MethodDelete, "/metadata/metaobject-definitions/"+definitionID, nil)
	deleteDefinitionReq.Header.Set(headers.IfMatch, quoteVersion(definitionVersion))
	deleteDefinitionRes, err := app.Test(deleteDefinitionReq, -1)
	if err != nil {
		t.Fatalf("delete definition Test() error = %v", err)
	}
	defer deleteDefinitionRes.Body.Close()
	if deleteDefinitionRes.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete definition StatusCode = %d, want %d", deleteDefinitionRes.StatusCode, fiber.StatusNoContent)
	}
}

// newTestApp creates a Fiber app with metadata routes.
func newTestApp(t *testing.T) *fiber.App {
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
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(headers.Middleware())
	app.Use(func(ctx *fiber.Ctx) error {
		principal.Set(ctx, principal.Principal{UserID: uuid.New(), SubjectHash: "test"})
		return ctx.Next()
	})
	v1 := app
	v1.Use(headers.RequireJSON())
	Register(v1, Services{Definitions: service, Values: service, Metaobjects: service, Checker: allowChecker{}})
	return app
}

// allowChecker permits route guards in HTTP adapter tests.
type allowChecker struct{}

// Check returns an allowed decision.
func (allowChecker) Check(context.Context, groupsport.CheckRequest) (groupsport.Decision, error) {
	return groupsport.Decision{Allowed: true, Reason: "test_allowed"}, nil
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
	req := httptest.NewRequest(fiber.MethodPost, "/metadata/metafield-definitions", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "create-definition")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	return res
}

// setOwnerValue writes the default owner value and returns its payload.
func setOwnerValue(t *testing.T, app *fiber.App, ownerID string, value string) map[string]any {
	t.Helper()
	body, err := json.Marshal(map[string]string{"value": value})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	req := httptest.NewRequest(fiber.MethodPut, "/metadata/owners/user/"+ownerID+"/metafields/profile/motto", bytes.NewReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "set-value-"+ownerID)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("set value Test() error = %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("set value StatusCode = %d, want %d", res.StatusCode, fiber.StatusCreated)
	}
	return decodeObject(t, res)
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
