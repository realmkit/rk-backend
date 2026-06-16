package metadata_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/api/headers"
)

// TestMetadataDefinitionLifecycle verifies metafield definition CRUD behavior.
func TestMetadataDefinitionLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)

	steps.Log("reject definition creation without idempotency")
	missingKey := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		`{"owner_type":"user","key":"motto","name":"Motto","value_type":"single_line_text"}`,
	)
	assertStatus(t, missingKey, fiber.StatusBadRequest)
	assertProblemCode(t, missingKey, "idempotency_key_required")

	steps.Log("create a user profile definition")
	created := createUserTextDefinition(t, fixture, "motto")
	id := idFrom(t, created, "id")
	version := versionFrom(t, created)

	steps.Log("read and list the definition")
	getResponse := fixture.doJSON(t, fiber.MethodGet, "/metadata/metafield-definitions/"+id.String(), "")
	assertStatus(t, getResponse, fiber.StatusOK)
	if getResponse.Header.Get(headers.ETag) == "" {
		t.Fatalf("%s header = empty", headers.ETag)
	}
	listResponse := fixture.doJSON(t, fiber.MethodGet, "/metadata/metafield-definitions?owner_type=user", "")
	assertStatus(t, listResponse, fiber.StatusOK)

	steps.Log("update definition with current version")
	patchResponse := fixture.doJSON(
		t,
		fiber.MethodPatch,
		"/metadata/metafield-definitions/"+id.String(),
		`{"name":"Public Motto","sort_order":3}`,
		withIfMatch(version),
	)
	assertStatus(t, patchResponse, fiber.StatusOK)
	updated := decodeObject(t, patchResponse)
	nextVersion := versionFrom(t, updated)
	if nextVersion <= version {
		t.Fatalf("updated version = %d, want greater than %d", nextVersion, version)
	}

	steps.Log("reject stale definition update")
	staleResponse := fixture.doJSON(
		t,
		fiber.MethodPatch,
		"/metadata/metafield-definitions/"+id.String(),
		`{"name":"Stale"}`,
		withIfMatch(version),
	)
	assertStatus(t, staleResponse, fiber.StatusPreconditionFailed)

	steps.Log("archive definition with latest version")
	deleteResponse := fixture.doJSON(
		t,
		fiber.MethodDelete,
		"/metadata/metafield-definitions/"+id.String(),
		"",
		withIfMatch(nextVersion),
	)
	assertStatus(t, deleteResponse, fiber.StatusNoContent)
}

// TestMetadataDefinitionValidationRejectsUnsupportedInput verifies definition validation.
func TestMetadataDefinitionValidationRejectsUnsupportedInput(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)

	steps.Log("reject unsupported owner type")
	unsupportedOwner := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		`{"owner_type":"forum_post","key":"motto","name":"Motto","value_type":"single_line_text"}`,
		withIdempotency("unsupported-owner"),
	)
	assertStatus(t, unsupportedOwner, fiber.StatusUnprocessableEntity)

	steps.Log("reject unsupported value type")
	unsupportedValue := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		`{"owner_type":"user","key":"motto","name":"Motto","value_type":"html"}`,
		withIdempotency("unsupported-value"),
	)
	assertStatus(t, unsupportedValue, fiber.StatusUnprocessableEntity)

	steps.Log("reject duplicate active definition key")
	createUserTextDefinition(t, fixture, "motto")
	duplicate := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		`{"owner_type":"user","key":"motto","name":"Motto","value_type":"single_line_text"}`,
		withIdempotency("duplicate-definition"),
	)
	assertStatus(t, duplicate, fiber.StatusConflict)
}
