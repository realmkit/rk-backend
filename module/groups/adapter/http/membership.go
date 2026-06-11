package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
)

// membershipRequest is the membership assign body.
type membershipRequest struct {
	Status           domain.MembershipStatus `json:"status"`
	AssignedByUserID *uuid.UUID              `json:"assigned_by_user_id"`
	AssignedReason   string                  `json:"assigned_reason"`
	StartsAt         *time.Time              `json:"starts_at"`
	ExpiresAt        *time.Time              `json:"expires_at"`
}

// membershipListResponse contains one membership page.
type membershipListResponse struct {
	Items         []domain.Membership `json:"items"`
	NextPageToken string              `json:"next_page_token,omitempty"`
}

// listGroupMembers lists memberships for a group.
func (handler handler) listGroupMembers(ctx *fiber.Ctx) error {
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Memberships.ListGroupMembers(ctx.UserContext(), groupID, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, membershipListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// assignMembership assigns a membership.
func (handler handler) assignMembership(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	groupID, userID, err := membershipIDs(ctx)
	if err != nil {
		return err
	}
	var request membershipRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	membership, err := handler.services.Memberships.Assign(
		ctx.UserContext(),
		port.AssignMembershipCommand{Membership: membershipFromRequest(groupID, userID, request)},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, membership.Version)
	return writeJSON(ctx, fiber.StatusOK, membership)
}

// removeMembership removes a membership.
func (handler handler) removeMembership(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	groupID, userID, err := membershipIDs(ctx)
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.RemoveMembershipCommand{
		GroupID:         groupID,
		UserID:          userID,
		ExpectedVersion: &version,
	}
	if err := handler.services.Memberships.Remove(ctx.UserContext(), command); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// listUserGroups lists groups for user.
func (handler handler) listUserGroups(ctx *fiber.Ctx) error {
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return err
	}
	return handler.writeUserGroups(ctx, userID)
}

// listCurrentUserGroups lists groups for current user.
func (handler handler) listCurrentUserGroups(ctx *fiber.Ctx) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	return handler.writeUserGroups(ctx, userID)
}

// writeUserGroups writes user group response.
func (handler handler) writeUserGroups(ctx *fiber.Ctx, userID uuid.UUID) error {
	groups, err := handler.services.Memberships.ListUserGroups(ctx.UserContext(), userID)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, groups)
}

// membershipIDs parses membership path IDs.
func membershipIDs(ctx *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return groupID, userID, nil
}

// membershipFromRequest maps HTTP request to membership.
func membershipFromRequest(groupID uuid.UUID, userID uuid.UUID, request membershipRequest) domain.Membership {
	return domain.Membership{
		GroupID:          groupID,
		UserID:           userID,
		Status:           request.Status,
		AssignedByUserID: request.AssignedByUserID,
		AssignedReason:   request.AssignedReason,
		StartsAt:         request.StartsAt,
		ExpiresAt:        request.ExpiresAt,
	}
}
