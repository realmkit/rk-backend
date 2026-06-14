package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Service manages local users.
type Service interface {
	// Get returns one user.
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)

	// Current returns the current user aggregate.
	Current(ctx context.Context, userID uuid.UUID) (CurrentUser, error)

	// List returns matching users.
	List(ctx context.Context, filter UserFilter, page pagination.Page) (pagination.Result[UserSummary], error)

	// FindSummariesByIDs returns display summaries for the requested local users.
	FindSummariesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]UserSummary, error)

	// UpdateCurrent updates local settings for the current user.
	UpdateCurrent(ctx context.Context, command UpdateCurrentCommand) (domain.User, error)
}
