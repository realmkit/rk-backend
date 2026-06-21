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
	Versions    port.VersionRepository         // Versions stores the versions value.
	Issues      port.ValidationIssueRepository // Issues stores the issues value.
	Signatures  port.SignatureRepository       // Signatures stores the signatures value.
	Activations port.ActivationRepository      // Activations stores the activations value.
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
	repositories Repositories      // repositories stores the repositories value.
	permissions  PermissionChecker // permissions stores the permissions value.
	events       EventSink         // events stores the events value.
	clock        Clock             // clock stores the clock value.
}

// ActivateCommand requests activation of one version.
type ActivateCommand struct {
	VersionID        uuid.UUID                    // VersionID stores the version i d value.
	Environment      domain.ActivationEnvironment // Environment stores the environment value.
	Reason           string                       // Reason stores the reason value.
	SettingsDataJSON []byte                       // SettingsDataJSON stores the settings data j s o n value.
	ActorUserID      *uuid.UUID                   // ActorUserID stores the actor user i d value.
}

// RollbackCommand requests activation of a previous version.
type RollbackCommand struct {
	ActivationID uuid.UUID                    // ActivationID stores the activation i d value.
	Environment  domain.ActivationEnvironment // Environment stores the environment value.
	Reason       string                       // Reason stores the reason value.
	ActorUserID  *uuid.UUID                   // ActorUserID stores the actor user i d value.
}
