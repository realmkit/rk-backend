package metadata_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/module/metadata/domain"
)

// TestMetadataSupportsForumOwners verifies planned forum owner types.
func TestMetadataSupportsForumOwners(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)
	owners := map[domain.OwnerType]uuid.UUID{
		domain.OwnerForumCategory: uuid.New(),
		domain.OwnerForum:         uuid.New(),
		domain.OwnerForumThread:   uuid.New(),
	}

	steps.Log("register forum owner fixtures")
	for ownerType, ownerID := range owners {
		fixture.owners.AddOwner(ownerType, ownerID)
	}

	steps.Log("write metadata on every forum owner type")
	for ownerType, ownerID := range owners {
		createOwnerLabelDefinition(t, fixture, ownerType)
		path := "/metadata/owners/" + string(ownerType) + "/" + ownerID.String() + "/metafields/display/label"
		response := fixture.doJSON(
			t,
			fiber.MethodPut,
			path,
			`{"value":"Featured"}`,
			withIdempotency("set-"+string(ownerType)),
		)
		assertStatus(t, response, fiber.StatusCreated)
		getResponse := fixture.doJSON(t, fiber.MethodGet, path, "")
		assertStatus(t, getResponse, fiber.StatusOK)
	}

	steps.Log("reject forum_post as an unsupported metadata owner")
	response := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		`{"owner_type":"forum_post","namespace":"display","key":"label","name":"Label","value_type":"single_line_text"}`,
		withIdempotency("forum-post-owner"),
	)
	assertStatus(t, response, fiber.StatusUnprocessableEntity)
}

// TestMetadataSupportsPunishmentAndTicketOwners verifies moderation owner types.
func TestMetadataSupportsPunishmentAndTicketOwners(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newMetadataFixture(t)
	owners := map[domain.OwnerType]uuid.UUID{
		domain.OwnerPunishmentDefinition: uuid.New(),
		domain.OwnerPunishment:           uuid.New(),
		domain.OwnerTicketDefinition:     uuid.New(),
		domain.OwnerTicket:               uuid.New(),
	}

	steps.Log("register punishment and ticket owners")
	for ownerType, ownerID := range owners {
		fixture.owners.AddOwner(ownerType, ownerID)
	}

	steps.Log("write display color through metadata")
	for ownerType, ownerID := range owners {
		createOwnerColorDefinition(t, fixture, ownerType)
		response := fixture.doJSON(
			t,
			fiber.MethodPut,
			"/metadata/owners/"+string(ownerType)+"/"+ownerID.String()+"/metafields/display/color",
			`{"value":"#3366ff"}`,
			withIdempotency("color-"+string(ownerType)),
		)
		assertStatus(t, response, fiber.StatusCreated)
	}
}

// createOwnerLabelDefinition creates a single-line owner label definition.
func createOwnerLabelDefinition(t *testing.T, fixture metadataFixture, ownerType domain.OwnerType) {
	t.Helper()
	body := `{"owner_type":"` + string(ownerType) + `","namespace":"display","key":"label","name":"Label","value_type":"single_line_text"}`
	response := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		body,
		withIdempotency("label-definition-"+string(ownerType)),
	)
	assertStatus(t, response, fiber.StatusCreated)
}

// createOwnerColorDefinition creates a color definition for owner presentation.
func createOwnerColorDefinition(t *testing.T, fixture metadataFixture, ownerType domain.OwnerType) {
	t.Helper()
	body := `{"owner_type":"` + string(ownerType) + `","namespace":"display","key":"color","name":"Color","value_type":"color"}`
	response := fixture.doJSON(
		t,
		fiber.MethodPost,
		"/metadata/metafield-definitions",
		body,
		withIdempotency("color-definition-"+string(ownerType)),
	)
	assertStatus(t, response, fiber.StatusCreated)
}
