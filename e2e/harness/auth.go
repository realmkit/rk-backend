package harness

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/identity"
)

// DevProvisioner creates principals for e2e development-auth requests.
type DevProvisioner struct{}

// Provision converts a validated identity into a test principal.
func (DevProvisioner) Provision(
	context.Context,
	identity.ExternalIdentity,
	auth.Token,
) (principal.Principal, error) {
	return principal.Principal{}, auth.ErrInvalidToken
}

// DevelopmentPrincipal creates a principal for the supplied local test user.
func (DevProvisioner) DevelopmentPrincipal(
	_ context.Context,
	userID uuid.UUID,
) (principal.Principal, error) {
	return principal.Principal{
		UserID:            userID,
		Issuer:            "realmkit-e2e",
		SubjectHash:       "e2e:" + userID.String(),
		Scopes:            []string{"e2e"},
		DevelopmentBypass: true,
	}, nil
}
