package groups_e2e

import (
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/e2e/harness"
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

// TestGroupsSearchAndCursorPagination verifies the searchable list contract.
func TestGroupsSearchAndCursorPagination(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)

	steps.Log("create searchable groups")
	builders := fixture.createGroup(t, "builders")
	moderators := fixture.createGroup(t, "moderators")

	steps.Log("search by group text query")
	searchResult := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/groups?q=builder&sort=key&direction=asc&page_size=10", ""))
	assertGroupsStatus(t, searchResult, fiber.StatusOK)
	searchPayload := decodeGroupsObject(t, searchResult)
	assertGroupListIDs(t, searchPayload, []string{builders["id"].(string)})
	if searchPayload["query"] != "builder" || searchPayload["sort"] != "key" || searchPayload["direction"] != "asc" {
		t.Fatalf("search metadata = %+v, want echoed query, sort, and direction", searchPayload)
	}

	steps.Log("page through groups with cursor token")
	firstPage := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/groups?sort=key&direction=asc&page_size=1", ""))
	assertGroupsStatus(t, firstPage, fiber.StatusOK)
	firstPayload := decodeGroupsObject(t, firstPage)
	assertGroupListIDs(t, firstPayload, []string{builders["id"].(string)})
	token, ok := firstPayload["next_page_token"].(string)
	if !ok || token == "" {
		t.Fatalf("next_page_token = %v, want cursor", firstPayload["next_page_token"])
	}

	secondPage := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/groups?sort=key&direction=asc&page_size=1&page_token="+url.QueryEscape(token), ""))
	assertGroupsStatus(t, secondPage, fiber.StatusOK)
	assertGroupListIDs(t, decodeGroupsObject(t, secondPage), []string{moderators["id"].(string)})
}

// assertGroupListIDs verifies group list item IDs in order.
func assertGroupListIDs(t *testing.T, payload map[string]any, expected []string) {
	t.Helper()
	items := payload["items"].([]any)
	if len(items) != len(expected) {
		t.Fatalf("items = %v, want %d items", items, len(expected))
	}
	for index, id := range expected {
		item := items[index].(map[string]any)
		if item["id"] != id {
			t.Fatalf("items[%d].id = %v, want %s", index, item["id"], id)
		}
	}
}
