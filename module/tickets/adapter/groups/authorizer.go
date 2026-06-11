// Package groups adapts the groups permission checker to ticket authorization.
package groups

import (
	"context"

	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
)

// Authorizer checks ticket permissions through the groups module.
type Authorizer struct {
	checker groupsport.Checker
}

// NewAuthorizer creates a ticket authorizer backed by groups permissions.
func NewAuthorizer(checker groupsport.Checker) Authorizer {
	return Authorizer{checker: checker}
}

// CanCreate reports whether actor can open tickets for a definition.
func (authorizer Authorizer) CanCreate(ctx context.Context, actorID uuid.UUID, definitionID uuid.UUID) (bool, error) {
	return authorizer.checkTicket(ctx, actorID, definitionID, groupsdomain.PermissionTicketsCreate)
}

// CanView reports whether actor can read a ticket.
func (authorizer Authorizer) CanView(ctx context.Context, actorID uuid.UUID, ticketID uuid.UUID) (bool, error) {
	return authorizer.checkTicket(ctx, actorID, ticketID, groupsdomain.PermissionTicketsView)
}

// CanReply reports whether actor can reply to a ticket.
func (authorizer Authorizer) CanReply(ctx context.Context, actorID uuid.UUID, ticketID uuid.UUID) (bool, error) {
	return authorizer.checkTicket(ctx, actorID, ticketID, groupsdomain.PermissionTicketsReply)
}

// CanStaffAction reports whether actor can perform staff ticket actions.
func (authorizer Authorizer) CanStaffAction(ctx context.Context, actorID uuid.UUID, ticketID uuid.UUID) (bool, error) {
	return authorizer.checkTicket(ctx, actorID, ticketID, groupsdomain.PermissionTicketsPerformActions)
}

// CanRevokePunishmentFromAppeal reports whether actor may revoke a punishment from an accepted appeal.
func (authorizer Authorizer) CanRevokePunishmentFromAppeal(
	ctx context.Context,
	actorID uuid.UUID,
	ticketID uuid.UUID,
	punishmentID uuid.UUID,
) (bool, error) {
	ticketAllowed, err := authorizer.checkTicket(ctx, actorID, ticketID, groupsdomain.PermissionTicketsPerformActions)
	if err != nil || !ticketAllowed {
		return ticketAllowed, err
	}
	return authorizer.checkPunishment(ctx, actorID, punishmentID, groupsdomain.PermissionPunishmentsRevoke)
}

func (authorizer Authorizer) checkTicket(
	ctx context.Context,
	actorID uuid.UUID,
	objectID uuid.UUID,
	permission groupsdomain.Permission,
) (bool, error) {
	return authorizer.check(ctx, actorID, groupsdomain.ObjectTicket, objectID, permission)
}

func (authorizer Authorizer) checkPunishment(
	ctx context.Context,
	actorID uuid.UUID,
	objectID uuid.UUID,
	permission groupsdomain.Permission,
) (bool, error) {
	return authorizer.check(ctx, actorID, groupsdomain.ObjectPunishment, objectID, permission)
}

func (authorizer Authorizer) check(
	ctx context.Context,
	actorID uuid.UUID,
	objectType groupsdomain.ObjectType,
	objectID uuid.UUID,
	permission groupsdomain.Permission,
) (bool, error) {
	if authorizer.checker == nil {
		return false, nil
	}
	decision, err := authorizer.checker.Check(ctx, groupsport.CheckRequest{
		ActorUserID: actorID,
		Permission:  permission,
		ObjectType:  objectType,
		ObjectID:    objectID,
	})
	return decision.Allowed, err
}
