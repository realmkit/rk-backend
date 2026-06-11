package port

import (
	"context"

	"github.com/google/uuid"
)

// RestrictionChecker denies actions blocked by active punishments.
type RestrictionChecker interface {
	// Restricted reports whether userID is denied actionKey.
	Restricted(ctx context.Context, userID uuid.UUID, actionKey string) (bool, error)
}
