package metadata_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestMetaobjectDefinitionLifecycle verifies metaobject definition behavior.
func TestMetaobjectDefinitionLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)

	steps.Log("create a profile card metaobject definition")
	createResponse := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metaobject-definitions",
		`{"type":"profile_card","name":"Profile Card","fields":[{"key":"motto","name":"Motto","value_type":"single_line_text","required":true}]}`,
		withIdempotency("create-profile-card"),
	)
	assertStatus(t, createResponse, fiber.StatusCreated)
	created := decodeObject(t, createResponse)
	id := idFrom(t, created, "id")
	version := versionFrom(t, created)

	steps.Log("read and list the metaobject definition")
	getResponse := fixture.doJSON(t, fiber.MethodGet, "/metadata/metaobject-definitions/"+id.String(), "")
	assertStatus(t, getResponse, fiber.StatusOK)
	listResponse := fixture.doJSON(t, fiber.MethodGet, "/metadata/metaobject-definitions?type=profile_card&active=true", "")
	assertStatus(t, listResponse, fiber.StatusOK)

	steps.Log("update definition display fields")
	patchResponse := fixture.doJSON(
		t,
		fiber.MethodPatch,
		"/metadata/metaobject-definitions/"+id.String(),
		`{"name":"Profile Card Updated","description":"Shown on profiles"}`,
		withIfMatch(version),
	)
	assertStatus(t, patchResponse, fiber.StatusOK)
	updated := decodeObject(t, patchResponse)
	nextVersion := versionFrom(t, updated)

	steps.Log("archive unused definition")
	deleteResponse := fixture.doJSON(
		t,
		fiber.MethodDelete,
		"/metadata/metaobject-definitions/"+id.String(),
		"",
		withIfMatch(nextVersion),
	)
	assertStatus(t, deleteResponse, fiber.StatusNoContent)
}

// TestMetaobjectEntryLifecycle verifies entry CRUD and validation.
func TestMetaobjectEntryLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)

	steps.Log("create metaobject definition")
	definition := createProfileCardDefinition(t, fixture)
	definitionID := idFrom(t, definition, "id")

	steps.Log("reject entry missing required field")
	invalid := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metaobject-definitions/"+definitionID.String()+"/entries",
		`{"handle":"first_card","display_name":"First Card","fields":{}}`,
		withIdempotency("invalid-entry"),
	)
	assertStatus(t, invalid, fiber.StatusUnprocessableEntity)

	steps.Log("create entry with required field")
	entryResponse := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metaobject-definitions/"+definitionID.String()+"/entries",
		`{"handle":"first_card","display_name":"First Card","fields":{"motto":"Ready"}}`,
		withIdempotency("valid-entry"),
	)
	assertStatus(t, entryResponse, fiber.StatusCreated)
	entry := decodeObject(t, entryResponse)
	entryID := idFrom(t, entry, "id")
	version := versionFrom(t, entry)
	fixture.owners.AddEntry(definitionID, entryID)

	steps.Log("list, read, update, and delete entry")
	listResponse := fixture.doJSON(t, fiber.MethodGet, "/metadata/metaobject-definitions/"+definitionID.String()+"/entries", "")
	assertStatus(t, listResponse, fiber.StatusOK)
	getResponse := fixture.doJSON(
		t,
		fiber.MethodGet,
		"/metadata/metaobject-definitions/"+definitionID.String()+"/entries/"+entryID.String(),
		"",
	)
	assertStatus(t, getResponse, fiber.StatusOK)
	patchResponse := fixture.doJSON(
		t,
		fiber.MethodPatch,
		"/metadata/metaobject-definitions/"+definitionID.String()+"/entries/"+entryID.String(),
		`{"display_name":"Updated Card","fields":{"motto":"Still ready"}}`,
		withIfMatch(version),
	)
	assertStatus(t, patchResponse, fiber.StatusOK)
	updated := decodeObject(t, patchResponse)
	deleteResponse := fixture.doJSON(
		t,
		fiber.MethodDelete,
		"/metadata/metaobject-definitions/"+definitionID.String()+"/entries/"+entryID.String(),
		"",
		withIfMatch(versionFrom(t, updated)),
	)
	assertStatus(t, deleteResponse, fiber.StatusNoContent)
}

// createProfileCardDefinition creates a reusable metaobject definition.
func createProfileCardDefinition(t *testing.T, fixture metadataFixture) map[string]any {
	t.Helper()
	response := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metaobject-definitions",
		`{"type":"profile_card","name":"Profile Card","fields":[{"key":"motto","name":"Motto","value_type":"single_line_text","required":true}]}`,
		withIdempotency("profile-card-definition"),
	)
	assertStatus(t, response, fiber.StatusCreated)
	return decodeObject(t, response)
}
