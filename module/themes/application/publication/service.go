package publication

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// NewService creates a publication service.
func NewService(
	repositories Repositories,
	permissions PermissionChecker,
	events EventSink,
	clock Clock,
) Service {
	if clock == nil {
		clock = time.Now
	}
	return Service{repositories: repositories, permissions: permissions, events: events, clock: clock}
}

// Activate activates a valid version for an environment.
func (service Service) Activate(ctx context.Context, command ActivateCommand) (domain.ThemeActivation, error) {
	if err := service.checkPermission(ctx, command.ActorUserID, command.Environment); err != nil {
		return domain.ThemeActivation{}, err
	}
	version, err := service.repositories.Versions.FindByID(ctx, command.VersionID)
	if err != nil {
		return domain.ThemeActivation{}, err
	}
	if err := service.ensurePublishable(ctx, version); err != nil {
		return domain.ThemeActivation{}, err
	}
	settings, err := activationSettings(version, command.SettingsDataJSON)
	if err != nil {
		return domain.ThemeActivation{}, err
	}
	activation := domain.ThemeActivation{
		ID:               uuid.New(),
		ThemeID:          version.ThemeID,
		VersionID:        version.ID,
		Environment:      command.Environment,
		Reason:           command.Reason,
		SettingsDataJSON: settings,
		ActivatedBy:      command.ActorUserID,
		ActivatedAt:      service.clock().UTC(),
	}
	if err := activation.Validate(); err != nil {
		return domain.ThemeActivation{}, err
	}
	activation, err = service.repositories.Activations.Activate(ctx, activation)
	if err != nil {
		return domain.ThemeActivation{}, err
	}
	if err := service.markPublished(ctx, version, command.ActorUserID); err != nil {
		return domain.ThemeActivation{}, err
	}
	return activation, service.emitActivated(ctx, activation)
}

// Rollback reactivates a previous activation target.
func (service Service) Rollback(ctx context.Context, command RollbackCommand) (domain.ThemeActivation, error) {
	if err := service.checkPermission(ctx, command.ActorUserID, command.Environment); err != nil {
		return domain.ThemeActivation{}, err
	}
	previous, err := service.repositories.Activations.FindByID(ctx, command.ActivationID)
	if err != nil {
		return domain.ThemeActivation{}, err
	}
	return service.Activate(ctx, ActivateCommand{
		VersionID:        previous.VersionID,
		Environment:      command.Environment,
		Reason:           command.Reason,
		SettingsDataJSON: previous.SettingsDataJSON,
		ActorUserID:      command.ActorUserID,
	})
}

// Current returns the current activation for an environment.
func (service Service) Current(
	ctx context.Context,
	environment domain.ActivationEnvironment,
) (domain.ThemeActivation, error) {
	return service.repositories.Activations.Current(ctx, environment)
}

// History returns activation history for a theme.
func (service Service) History(ctx context.Context, themeID uuid.UUID) ([]domain.ThemeActivation, error) {
	return service.repositories.Activations.ListByTheme(ctx, themeID)
}

// checkPermission verifies activation permission.
func (service Service) checkPermission(
	ctx context.Context,
	actorUserID *uuid.UUID,
	environment domain.ActivationEnvironment,
) error {
	if service.permissions == nil {
		return nil
	}
	return service.permissions.CanActivate(ctx, actorUserID, environment)
}

// emitActivated emits cache and audit events.
func (service Service) emitActivated(ctx context.Context, activation domain.ThemeActivation) error {
	if service.events == nil {
		return nil
	}
	return service.events.ThemeActivated(ctx, activation)
}
