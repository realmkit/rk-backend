package groups_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestGroupsEmitLifecycleEvents verifies group, membership, and tuple event facts.
func TestGroupsEmitLifecycleEvents(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()

	steps.Log("create, update, and delete a group")
	group := fixture.createGroup(t, "events")
	groupID := groupIDFrom(t, group)
	updated := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/groups/"+groupID.String(), groupUpdateBody()),
			withGroupsIdempotency("events-update-group"),
			withGroupsIfMatch(groupVersionFrom(t, group)),
		),
	)
	assertGroupsStatus(t, updated, fiber.StatusOK)
	updatedPayload := decodeGroupsObject(t, updated)

	steps.Log("assign and remove membership")
	assigned := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(
				fiber.MethodPut,
				"/groups/"+groupID.String()+"/members/"+userID.String(),
				`{"status":"active","assigned_reason":"events"}`,
			),
			withGroupsIdempotency("events-assign-member"),
		),
	)
	assertGroupsStatus(t, assigned, fiber.StatusOK)
	member := decodeGroupsObject(t, assigned)
	removed := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/groups/"+groupID.String()+"/members/"+userID.String(), ""),
			withGroupsIdempotency("events-remove-member"),
			withGroupsIfMatch(uint64(member["version"].(float64))),
		),
	)
	assertGroupsStatus(t, removed, fiber.StatusNoContent)

	steps.Log("create and delete a relation tuple")
	tuple := fixture.createTuple(
		t,
		domain.RelationTuple{
			ObjectType:  domain.ObjectForum,
			ObjectID:    uuid.New(),
			Relation:    domain.RelationViewer,
			SubjectType: domain.SubjectUser,
			SubjectID:   userID,
		},
	)
	if err := fixture.service.DeleteTuple(context.Background(), port.DeleteTupleCommand{ID: tuple.ID}); err != nil {
		t.Fatalf("DeleteTuple() error = %v", err)
	}

	steps.Log("delete group")
	deleted := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/groups/"+groupID.String(), ""),
			withGroupsIdempotency("events-delete-group"),
			withGroupsIfMatch(groupVersionFrom(t, updatedPayload)),
		),
	)
	assertGroupsStatus(t, deleted, fiber.StatusNoContent)

	steps.Log("verify emitted event keys")
	drafts := fixture.events.Drafts()
	for _, key := range []eventdomain.EventKey{
		"groups.group.created",
		"groups.group.updated",
		eventdomain.EventGroupsMembershipAdded,
		"groups.membership.removed",
		"groups.relation_tuple.created",
		"groups.relation_tuple.deleted",
		"groups.group.deleted",
	} {
		assertGroupsEventPresent(t, drafts, key)
	}
}

// TestGroupsOpenAPICoversRoutes verifies group route contract coverage.
func TestGroupsOpenAPICoversRoutes(t *testing.T) {
	steps := harness.NewSteps(t)
	routes := []struct {
		method string
		path   string
	}{
		{fiber.MethodPost, "/groups"},
		{fiber.MethodGet, "/groups"},
		{fiber.MethodGet, "/groups/{group_id}"},
		{fiber.MethodPatch, "/groups/{group_id}"},
		{fiber.MethodDelete, "/groups/{group_id}"},
		{fiber.MethodGet, "/groups/{group_id}/members"},
		{fiber.MethodPut, "/groups/{group_id}/members/{user_id}"},
		{fiber.MethodDelete, "/groups/{group_id}/members/{user_id}"},
		{fiber.MethodGet, "/users/{user_id}/groups"},
		{fiber.MethodGet, "/users/me/groups"},
		{fiber.MethodPost, "/permissions/check"},
	}
	for _, route := range routes {
		steps.Log("verify OpenAPI operation %s %s", route.method, route.path)
		assertGroupsOpenAPIRoute(t, route.method, route.path)
	}
}

// groupUpdateBody returns a valid group update body.
func groupUpdateBody() string {
	return `{"name":"Events Updated","description":"Updated","color":"#3366ff","weight":70,"status":"active"}`
}

// assertGroupsEventPresent verifies one group event key was published.
func assertGroupsEventPresent(t *testing.T, drafts []eventdomain.Draft, key eventdomain.EventKey) {
	t.Helper()
	for _, draft := range drafts {
		if draft.Key == key {
			if draft.Producer != eventdomain.ProducerGroups {
				t.Fatalf("producer = %s, want %s", draft.Producer, eventdomain.ProducerGroups)
			}
			if len(draft.Scopes) == 0 {
				t.Fatalf("event %s has no scopes", key)
			}
			return
		}
	}
	t.Fatalf("event %s missing from %+v", key, drafts)
}
