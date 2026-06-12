// Package authgate exposes small helpers for route-level authentication gates.
package authgate

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/api/problem"
)

// RequirePrincipal returns the authenticated principal or a problem error.
func RequirePrincipal(ctx *fiber.Ctx) (principal.Principal, error) {
	current, err := principal.Require(ctx)
	if err != nil {
		return principal.Principal{}, problem.Error{
			Problem: problem.New(fiber.StatusUnauthorized, "unauthenticated", "Authentication is required."),
		}
	}
	return current, nil
}

// OptionalPrincipal returns the authenticated principal when one is present.
func OptionalPrincipal(ctx *fiber.Ctx) (principal.Principal, bool) {
	return principal.Current(ctx)
}

// RequireUserID returns the authenticated local user ID.
func RequireUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	current, err := RequirePrincipal(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return current.UserID, nil
}

// OptionalUserID returns the authenticated user ID when one is present.
func OptionalUserID(ctx *fiber.Ctx) uuid.UUID {
	current, ok := OptionalPrincipal(ctx)
	if !ok {
		return uuid.Nil
	}
	return current.UserID
}

// RequireScope verifies that the current principal has one OAuth scope.
func RequireScope(ctx *fiber.Ctx, scope string) error {
	return RequireAnyScope(ctx, scope)
}

// RequireAnyScope verifies that the current principal has at least one scope.
func RequireAnyScope(ctx *fiber.Ctx, scopes ...string) error {
	current, err := RequirePrincipal(ctx)
	if err != nil {
		return err
	}
	for _, scope := range scopes {
		if slices.Contains(current.Scopes, scope) {
			return nil
		}
	}
	return problem.Error{
		Problem: problem.New(fiber.StatusForbidden, "insufficient_scope", "The access token does not grant this operation."),
	}
}
