// Package httpguard adapts group permission checks to Fiber handlers.
package httpguard

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/authgate"
	"github.com/realmkit/rk-backend/pkg/api/problem"
)

// allGuardScopeID is a concrete resource ID used to test all-scope grants.
var allGuardScopeID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// Target identifies one protected permission target.
type Target struct {
	// Action is the action being guarded.
	Action domain.Action

	// ScopeType is the target resource type.
	ScopeType domain.ScopeType

	// ScopeID is the target resource ID.
	ScopeID uuid.UUID
}

// All returns a target applying to all resources of one scope type.
func All(action domain.Action, scopeType domain.ScopeType) Target {
	return Target{Action: action, ScopeType: scopeType, ScopeID: allGuardScopeID}
}

// Object returns a target applying to one resource.
func Object(action domain.Action, scopeType domain.ScopeType, scopeID uuid.UUID) Target {
	return Target{Action: action, ScopeType: scopeType, ScopeID: scopeID}
}

// Require verifies the authenticated actor has a permission target.
func Require(ctx *fiber.Ctx, checker port.Checker, target Target) (uuid.UUID, error) {
	actor, err := authgate.RequireUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return actor, Check(ctx, checker, actor, target)
}

// RequireSelfOr verifies self access or a permission target.
func RequireSelfOr(ctx *fiber.Ctx, checker port.Checker, subject uuid.UUID, target Target) (uuid.UUID, error) {
	actor, err := authgate.RequireUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	if actor == subject {
		return actor, nil
	}
	return actor, Check(ctx, checker, actor, target)
}

// Check verifies one known actor against a permission target.
func Check(ctx *fiber.Ctx, checker port.Checker, actor uuid.UUID, target Target) error {
	if checker == nil {
		return problem.Error{
			Problem: problem.New(fiber.StatusInternalServerError, "authorization_unavailable", "Authorization is not configured."),
		}
	}
	decision, err := checker.Check(ctx.UserContext(), port.CheckRequest{
		ActorUserID: actor,
		Action:      target.Action,
		ScopeType:   target.ScopeType,
		ScopeID:     target.ScopeID,
	})
	if err != nil {
		if errors.Is(err, port.ErrUnknownPermission) {
			return problem.Error{
				Problem: problem.New(fiber.StatusInternalServerError, "authorization_unknown_permission", "Permission is not configured."),
			}
		}
		return err
	}
	if !decision.Allowed {
		return Denied()
	}
	return nil
}

// Denied returns the shared permission-denied problem.
func Denied() error {
	return problem.Error{
		Problem: problem.New(fiber.StatusForbidden, "permission_denied", "Permission was denied."),
	}
}
