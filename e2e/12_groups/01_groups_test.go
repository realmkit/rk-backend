package groups_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestGroupsLifecycle verifies group create, read, update, list, and delete.
func TestGroupsLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)

	steps.Log("reject group creation without idempotency")
	missingKey := fixture.do(
		t,
		harness.JSONRequest(fiber.MethodPost, "/groups", `{"key":"staff","name":"Staff","color":"#3366ff","status":"active"}`),
	)
	assertGroupsStatus(t, missingKey, fiber.StatusBadRequest)

	steps.Log("create group")
	created := fixture.createGroup(t, "staff")
	groupID := groupIDFrom(t, created)
	version := groupVersionFrom(t, created)

	steps.Log("fetch group by id")
	fetched := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/groups/"+groupID.String(), ""))
	assertGroupsStatus(t, fetched, fiber.StatusOK)
	if fetched.Header.Get("ETag") == "" {
		t.Fatalf("ETag missing from get group response")
	}

	steps.Log("list active groups")
	listed := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/groups?status=active&page_size=10", ""))
	assertGroupsStatus(t, listed, fiber.StatusOK)
	listPayload := decodeGroupsObject(t, listed)
	if len(listPayload["items"].([]any)) != 1 {
		t.Fatalf("items = %v, want one group", listPayload["items"])
	}

	steps.Log("reject update without If-Match")
	missingVersion := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/groups/"+groupID.String(), `{"name":"Staff Updated","color":"#6633ff","weight":60,"status":"active"}`),
			withGroupsIdempotency("missing-version"),
		),
	)
	assertGroupsStatus(t, missingVersion, fiber.StatusPreconditionRequired)

	steps.Log("update mutable group fields")
	updated := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/groups/"+groupID.String(), `{"name":"Staff Updated","color":"#6633ff","weight":60,"status":"active"}`),
			withGroupsIdempotency("update-staff"),
			withGroupsIfMatch(version),
		),
	)
	assertGroupsStatus(t, updated, fiber.StatusOK)
	updatedPayload := decodeGroupsObject(t, updated)
	if updatedPayload["name"] != "Staff Updated" {
		t.Fatalf("name = %v, want Staff Updated", updatedPayload["name"])
	}

	steps.Log("reject stale update")
	stale := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/groups/"+groupID.String(), `{"name":"Stale","color":"#6633ff","weight":60,"status":"active"}`),
			withGroupsIdempotency("stale-staff"),
			withGroupsIfMatch(version),
		),
	)
	assertGroupsStatus(t, stale, fiber.StatusPreconditionFailed)

	steps.Log("delete group")
	deleted := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/groups/"+groupID.String(), ""),
			withGroupsIdempotency("delete-staff"),
			withGroupsIfMatch(groupVersionFrom(t, updatedPayload)),
		),
	)
	assertGroupsStatus(t, deleted, fiber.StatusNoContent)
}

// TestGroupsValidationRejectsInvalidInput verifies group validation and conflict mapping.
func TestGroupsValidationRejectsInvalidInput(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)

	steps.Log("reject invalid key and color")
	invalid := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/groups", `{"key":"Bad Key","name":"","color":"blue","status":"active"}`),
			withGroupsIdempotency("invalid-group"),
		),
	)
	assertGroupsStatus(t, invalid, fiber.StatusUnprocessableEntity)

	steps.Log("create original group")
	fixture.createGroup(t, "builders")

	steps.Log("reject duplicate active key")
	duplicate := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/groups", `{"key":"builders","name":"Builders 2","color":"#3366ff","status":"active"}`),
			withGroupsIdempotency("duplicate-builders"),
		),
	)
	assertGroupsStatus(t, duplicate, fiber.StatusConflict)
}
