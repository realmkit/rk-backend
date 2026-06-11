package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
)

// Service manages local users.
type Service interface {
	// Get returns one user.
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)

	// Current returns the current user aggregate.
	Current(ctx context.Context, userID uuid.UUID) (CurrentUser, error)

	// UpdateCurrent updates local settings for the current user.
	UpdateCurrent(ctx context.Context, command UpdateCurrentCommand) (domain.User, error)
}
