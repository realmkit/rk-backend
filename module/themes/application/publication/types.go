package publication

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by publication.
type Repositories struct {
	Versions    port.VersionRepository
	Issues      port.ValidationIssueRepository
	Signatures  port.SignatureRepository
	Activations port.ActivationRepository
}

// PermissionChecker verifies publication permissions.
type PermissionChecker interface {
	CanActivate(context.Context, *uuid.UUID, domain.ActivationEnvironment) error
}

// EventSink receives publication events.
type EventSink interface {
	ThemeActivated(context.Context, domain.ThemeActivation) error
}

// Clock returns the current time.
type Clock func() time.Time

// Service owns theme activation, publication, and rollback workflows.
type Service struct {
	repositories Repositories
	permissions  PermissionChecker
	events       EventSink
	clock        Clock
}

// ActivateCommand requests activation of one version.
type ActivateCommand struct {
	VersionID        uuid.UUID
	Environment      domain.ActivationEnvironment
	Reason           string
	SettingsDataJSON []byte
	ActorUserID      *uuid.UUID
}

// RollbackCommand requests activation of a previous version.
type RollbackCommand struct {
	ActivationID uuid.UUID
	Environment  domain.ActivationEnvironment
	Reason       string
	ActorUserID  *uuid.UUID
}
