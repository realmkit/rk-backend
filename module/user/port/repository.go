package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// UserRepository stores local users.
type UserRepository interface {
	// Create stores a user.
	Create(ctx context.Context, user domain.User) (domain.User, error)

	// Update stores mutable user fields.
	Update(ctx context.Context, user domain.User, expectedVersion uint64) (domain.User, error)

	// FindByID returns one user.
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)

	// List returns matching users.
	List(ctx context.Context, filter UserFilter, page pagination.Page) (pagination.Result[UserSummary], error)

	// FindSummariesByIDs returns display summaries keyed by local user ID.
	FindSummariesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]UserSummary, error)

	// TouchLastSeen stores the last-seen timestamp.
	TouchLastSeen(ctx context.Context, id uuid.UUID) error
}

// IdentityLinkRepository stores identity links.
type IdentityLinkRepository interface {
	// Create stores an identity link.
	Create(ctx context.Context, link domain.IdentityLink) (domain.IdentityLink, error)

	// FindByIssuerSubject returns one identity link.
	FindByIssuerSubject(ctx context.Context, issuer string, subject string) (domain.IdentityLink, error)

	// Touch stores last-seen and sync information.
	Touch(ctx context.Context, link domain.IdentityLink) error
}

// ClaimCacheRepository stores provider claim caches.
type ClaimCacheRepository interface {
	// Upsert stores provider claim cache data.
	Upsert(ctx context.Context, claims domain.ClaimCache) (domain.ClaimCache, error)

	// FindByUserID returns provider claim cache for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) (domain.ClaimCache, error)
}
