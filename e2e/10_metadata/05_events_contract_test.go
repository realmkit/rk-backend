package metadata_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/pkg/api/openapi"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestMetadataEmitsDefinitionValueAndMetaobjectEvents verifies event facts.
func TestMetadataEmitsDefinitionValueAndMetaobjectEvents(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)
	ownerID := uuid.New()
	fixture.owners.AddOwner(domain.OwnerUser, ownerID)

	steps.Log("create definition and set value")
	createUserTextDefinition(t, fixture, "motto")
	valueResponse := fixture.doJSON(
		t,
		fiber.MethodPut,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/motto",
		`{"value":"Ready"}`,
		withIdempotency("event-value"),
	)
	assertStatus(t, valueResponse, fiber.StatusCreated)
	value := decodeObject(t, valueResponse)

	steps.Log("delete value and create metaobject entry")
	deleteResponse := fixture.doJSON(
		t,
		fiber.MethodDelete,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/motto",
		"",
		withIfMatch(versionFrom(t, value)),
	)
	assertStatus(t, deleteResponse, fiber.StatusNoContent)
	definition := createProfileCardDefinition(t, fixture)
	definitionID := idFrom(t, definition, "id")
	entryResponse := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metaobject-definitions/"+definitionID.String()+"/entries",
		`{"handle":"event_card","display_name":"Event Card","fields":{"motto":"Ready"}}`,
		withIdempotency("event-entry"),
	)
	assertStatus(t, entryResponse, fiber.StatusCreated)

	steps.Log("verify safe event facts were recorded")
	drafts := fixture.events.Drafts()
	if len(drafts) < 5 {
		t.Fatalf("recorded events = %d, want at least 5", len(drafts))
	}
	assertEventKeyPresent(t, drafts, eventdomain.EventMetadataMetafieldSet)
	assertEventKeyPresent(t, drafts, "metadata.metafield.deleted")
	assertEventKeyPresent(t, drafts, "metadata.entry.created")
}

// TestMetadataOpenAPICoversRoutes verifies every metadata route is documented.
func TestMetadataOpenAPICoversRoutes(t *testing.T) {
	steps := harness.NewSteps(t)
	routes := []struct {
		method string
		path   string
	}{
		{fiber.MethodPost, "/metadata/metafield-definitions"},
		{fiber.MethodGet, "/metadata/metafield-definitions"},
		{fiber.MethodGet, "/metadata/metafield-definitions/{definition_id}"},
		{fiber.MethodPatch, "/metadata/metafield-definitions/{definition_id}"},
		{fiber.MethodDelete, "/metadata/metafield-definitions/{definition_id}"},
		{fiber.MethodPut, "/metadata/owners/{owner_type}/{owner_id}/metafields/{namespace}/{key}"},
		{fiber.MethodGet, "/metadata/owners/{owner_type}/{owner_id}/metafields"},
		{fiber.MethodGet, "/metadata/owners/{owner_type}/{owner_id}/metafields/{namespace}/{key}"},
		{fiber.MethodDelete, "/metadata/owners/{owner_type}/{owner_id}/metafields/{namespace}/{key}"},
		{fiber.MethodPost, "/metadata/metaobject-definitions"},
		{fiber.MethodGet, "/metadata/metaobject-definitions"},
		{fiber.MethodGet, "/metadata/metaobject-definitions/{definition_id}"},
		{fiber.MethodPatch, "/metadata/metaobject-definitions/{definition_id}"},
		{fiber.MethodDelete, "/metadata/metaobject-definitions/{definition_id}"},
		{fiber.MethodPost, "/metadata/metaobject-definitions/{definition_id}/entries"},
		{fiber.MethodGet, "/metadata/metaobject-definitions/{definition_id}/entries"},
		{fiber.MethodGet, "/metadata/metaobject-definitions/{definition_id}/entries/{entry_id}"},
		{fiber.MethodPatch, "/metadata/metaobject-definitions/{definition_id}/entries/{entry_id}"},
		{fiber.MethodDelete, "/metadata/metaobject-definitions/{definition_id}/entries/{entry_id}"},
	}

	for _, route := range routes {
		steps.Log("verify OpenAPI operation %s %s", route.method, route.path)
		ok, err := openapi.OperationExists(route.method, route.path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !ok {
			t.Fatalf("%s %s missing OpenAPI operation", route.method, route.path)
		}
	}
}

// assertEventKeyPresent verifies one event key was recorded.
func assertEventKeyPresent(t *testing.T, drafts []eventdomain.Draft, key eventdomain.EventKey) {
	t.Helper()
	for _, draft := range drafts {
		if draft.Key == key {
			if draft.Producer != eventdomain.ProducerMetadata {
				t.Fatalf("producer = %s, want %s", draft.Producer, eventdomain.ProducerMetadata)
			}
			if len(draft.Scopes) == 0 {
				t.Fatalf("event %s has no scopes", key)
			}
			return
		}
	}
	t.Fatalf("event %s missing from %+v", key, drafts)
}
