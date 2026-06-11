package metadata_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
)

// TestMetadataValueLifecycle verifies owner metafield writes and reads.
func TestMetadataValueLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)
	ownerID := uuid.New()
	fixture.owners.AddOwner(domain.OwnerUser, ownerID)

	steps.Log("create a definition for an existing owner")
	createUserTextDefinition(t, fixture, "motto")

	steps.Log("set a value for the owner")
	setResponse := fixture.doJSON(
		t,
		fiber.MethodPut,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/motto",
		`{"value":"Ready"}`,
		withIdempotency("set-motto"),
	)
	assertStatus(t, setResponse, fiber.StatusCreated)
	value := decodeObject(t, setResponse)
	version := versionFrom(t, value)

	steps.Log("get and list the owner value")
	getResponse := fixture.doJSON(
		t,
		fiber.MethodGet,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/motto",
		"",
	)
	assertStatus(t, getResponse, fiber.StatusOK)
	listResponse := fixture.doJSON(
		t,
		fiber.MethodGet,
		"/metadata/owners/user/"+ownerID.String()+"/metafields?namespace=profile&include_empty=false",
		"",
	)
	assertStatus(t, listResponse, fiber.StatusOK)

	steps.Log("update the value with optimistic concurrency")
	updateResponse := fixture.doJSON(
		t,
		fiber.MethodPut,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/motto",
		`{"value":"Still ready"}`,
		withIdempotency("update-motto"),
		withIfMatch(version),
	)
	assertStatus(t, updateResponse, fiber.StatusOK)
	updated := decodeObject(t, updateResponse)
	nextVersion := versionFrom(t, updated)

	steps.Log("delete the value")
	deleteResponse := fixture.doJSON(
		t,
		fiber.MethodDelete,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/motto",
		"",
		withIfMatch(nextVersion),
	)
	assertStatus(t, deleteResponse, fiber.StatusNoContent)
}

// TestMetadataValueValidationUsesDefinitionConstraints verifies value rules.
func TestMetadataValueValidationUsesDefinitionConstraints(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)
	ownerID := uuid.New()
	fixture.owners.AddOwner(domain.OwnerUser, ownerID)

	steps.Log("create constrained enum definition")
	body := `{"owner_type":"user","namespace":"profile","key":"rank","name":"Rank","value_type":"enum","rules":{"allowed_values":["member","mod"]}}`
	response := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		body,
		withIdempotency("rank-definition"),
	)
	assertStatus(t, response, fiber.StatusCreated)

	steps.Log("reject invalid enum value")
	invalid := fixture.doJSON(
		t,
		fiber.MethodPut,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/rank",
		`{"value":"owner"}`,
		withIdempotency("invalid-rank"),
	)
	assertStatus(t, invalid, fiber.StatusUnprocessableEntity)

	steps.Log("accept valid enum value")
	valid := fixture.doJSON(
		t,
		fiber.MethodPut,
		"/metadata/owners/user/"+ownerID.String()+"/metafields/profile/rank",
		`{"value":"member"}`,
		withIdempotency("valid-rank"),
	)
	assertStatus(t, valid, fiber.StatusCreated)
}

// TestMetadataValueRequiresExistingOwner verifies owner existence enforcement.
func TestMetadataValueRequiresExistingOwner(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)
	missingOwnerID := uuid.New()

	steps.Log("create definition without registering the target owner")
	createUserTextDefinition(t, fixture, "motto")

	steps.Log("reject value write for missing owner")
	response := fixture.doJSON(
		t,
		fiber.MethodPut,
		"/metadata/owners/user/"+missingOwnerID.String()+"/metafields/profile/motto",
		`{"value":"Ready"}`,
		withIdempotency("missing-owner-value"),
	)
	assertStatus(t, response, fiber.StatusNotFound)
}
