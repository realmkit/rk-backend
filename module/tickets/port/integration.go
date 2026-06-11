package port

import (
	"context"

	"github.com/google/uuid"
)

// PunishmentSummary is a safe punishment projection for ticket intake.
type PunishmentSummary struct {
	ID           uuid.UUID
	TargetUserID uuid.UUID
	IssuerUserID *uuid.UUID
	Status       string
}

// PunishmentReader resolves punishments for appeals.
type PunishmentReader interface {
	GetPunishment(context.Context, uuid.UUID) (PunishmentSummary, error)
}

// PunishmentExecutor performs punishment effects from ticket actions.
type PunishmentExecutor interface {
	RevokePunishment(context.Context, uuid.UUID, uuid.UUID, string, uint64) error
}

// AssetResolver checks evidence assets.
type AssetResolver interface {
	AssetExists(context.Context, uuid.UUID) (bool, error)
}

// UserResolver checks target and assignee users.
type UserResolver interface {
	UserExists(context.Context, uuid.UUID) (bool, error)
}

// GroupResolver checks staff team groups.
type GroupResolver interface {
	GroupExists(context.Context, uuid.UUID) (bool, error)
}

// Authorizer checks ticket permissions.
type Authorizer interface {
	CanCreate(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CanView(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CanReply(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CanStaffAction(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CanRevokePunishmentFromAppeal(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error)
}
