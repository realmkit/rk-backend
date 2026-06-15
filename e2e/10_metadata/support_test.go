// Package metadata_e2e verifies metadata journeys through the real server.
package metadata_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	metadatahttp "github.com/realmkit/rk-backend/module/metadata/adapter/http"
	metadatapostgres "github.com/realmkit/rk-backend/module/metadata/adapter/postgres"
	metadataapplication "github.com/realmkit/rk-backend/module/metadata/application"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/server"
)

// metadataFixture contains the metadata e2e module wiring.
type metadataFixture struct {
	ecosystem *harness.Ecosystem
	events    *eventtesting.PublisherRecorder
	owners    *knownOwners
	actorID   uuid.UUID
}

// knownOwners is an e2e owner and reference resolver.
type knownOwners struct {
	owners  map[string]struct{}
	entries map[string]struct{}
}

// newMetadataFixture starts a server with metadata routes.
func newMetadataFixture(t *testing.T) metadataFixture {
	t.Helper()
	owners := newKnownOwners()
	events := &eventtesting.PublisherRecorder{}
	database := harness.NewSQLiteDatabase(t)
	service := metadataapplication.NewService(metadataapplication.Dependencies{
		Definitions:           metadatapostgres.NewMetafieldDefinitionRepository(database.Store),
		Values:                metadatapostgres.NewMetafieldValueRepository(database.Store),
		MetaobjectDefinitions: metadatapostgres.NewMetaobjectDefinitionRepository(database.Store),
		MetaobjectEntries:     metadatapostgres.NewMetaobjectEntryRepository(database.Store),
		Owners:                owners,
		References:            owners,
		Events:                events,
	})
	actorID := uuid.MustParse("00000000-0000-0000-0000-00000000e010")
	ecosystem := harness.New(
		t,
		harness.WithDevelopment(true),
		harness.WithDatabase(database),
		harness.WithServerOptions(
			server.WithAuth(auth.Config{DevelopmentBypass: true}, harness.DevProvisioner{}),
			server.WithMetadata(metadatahttp.Services{
				Definitions: service,
				Values:      service,
				Metaobjects: service,
				Checker:     harness.AllowChecker{},
			}),
		),
	)
	return metadataFixture{ecosystem: ecosystem, events: events, owners: owners, actorID: actorID}
}

// newKnownOwners creates an empty owner resolver.
func newKnownOwners() *knownOwners {
	return &knownOwners{
		owners:  make(map[string]struct{}),
		entries: make(map[string]struct{}),
	}
}

// AddOwner registers an owner as existing.
func (resolver *knownOwners) AddOwner(ownerType domain.OwnerType, ownerID uuid.UUID) {
	resolver.owners[string(ownerType)+":"+ownerID.String()] = struct{}{}
}

// AddEntry registers a metaobject entry as existing.
func (resolver *knownOwners) AddEntry(definitionID uuid.UUID, entryID uuid.UUID) {
	resolver.entries[definitionID.String()+":"+entryID.String()] = struct{}{}
}

// Exists reports whether owner exists.
func (resolver *knownOwners) Exists(_ context.Context, ownerType domain.OwnerType, ownerID uuid.UUID) (bool, error) {
	_, ok := resolver.owners[string(ownerType)+":"+ownerID.String()]
	return ok, nil
}

// OwnerExists reports whether owner reference exists.
func (resolver *knownOwners) OwnerExists(_ context.Context, reference domain.OwnerReference) (bool, error) {
	_, ok := resolver.owners[string(reference.Type)+":"+reference.ID.String()]
	return ok, nil
}

// MetaobjectEntryExists reports whether metaobject reference exists.
func (resolver *knownOwners) MetaobjectEntryExists(_ context.Context, reference domain.MetaobjectReference) (bool, error) {
	_, ok := resolver.entries[reference.DefinitionID.String()+":"+reference.EntryID.String()]
	return ok, nil
}

// doJSON sends one JSON request.
func (fixture metadataFixture) doJSON(
	t *testing.T,
	method string,
	path string,
	body string,
	configure ...func(*http.Request),
) *http.Response {
	t.Helper()
	request := harness.JSONRequest(method, path, body)
	request.Header.Set(auth.DevUserIDHeader, fixture.actorID.String())
	for _, fn := range configure {
		fn(request)
	}
	return fixture.ecosystem.Test(t, request)
}

// withIdempotency adds an idempotency key.
func withIdempotency(key string) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IdempotencyKey, key)
	}
}

// withIfMatch adds an If-Match version header.
func withIfMatch(version uint64) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IfMatch, `"`+strconv.FormatUint(version, 10)+`"`)
	}
}

// decodeObject decodes a JSON object response.
func decodeObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// assertStatus verifies a status code.
func assertStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

// assertProblemCode verifies a problem code is present.
func assertProblemCode(t *testing.T, response *http.Response, code string) {
	t.Helper()
	payload := decodeObject(t, response)
	if payload["code"] != code {
		t.Fatalf("problem code = %v, want %s payload = %+v", payload["code"], code, payload)
	}
}

// versionFrom returns the numeric version from response payload.
func versionFrom(t *testing.T, payload map[string]any) uint64 {
	t.Helper()
	value, ok := payload["version"].(float64)
	if !ok {
		t.Fatalf("version missing in payload %+v", payload)
	}
	return uint64(value)
}

// idFrom returns a UUID field from response payload.
func idFrom(t *testing.T, payload map[string]any, field string) uuid.UUID {
	t.Helper()
	raw, ok := payload[field].(string)
	if !ok {
		t.Fatalf("%s missing in payload %+v", field, payload)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("uuid.Parse(%s) error = %v", raw, err)
	}
	return id
}

// createUserTextDefinition creates the default user profile definition.
func createUserTextDefinition(t *testing.T, fixture metadataFixture, key string) map[string]any {
	t.Helper()
	body := `{"owner_type":"user","namespace":"profile","key":"` + key + `","name":"Profile ` + key + `","value_type":"single_line_text","rules":{"max_length":80}}`
	response := fixture.doJSON(t, fiber.MethodPost, "/metadata/metafield-definitions", body, withIdempotency("definition-"+key))
	assertStatus(t, response, fiber.StatusCreated)
	return decodeObject(t, response)
}
