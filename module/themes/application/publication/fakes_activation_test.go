package publication

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeActivationRepository stores activation history.
type fakeActivationRepository struct {
	activations map[uuid.UUID]domain.ThemeActivation
}

// Activate stores one activation.
func (repository *fakeActivationRepository) Activate(
	_ context.Context,
	activation domain.ThemeActivation,
) (domain.ThemeActivation, error) {
	for id, existing := range repository.activations {
		if existing.Environment == activation.Environment && existing.IsCurrent {
			existing.IsCurrent = false
			repository.activations[id] = existing
		}
	}
	activation.IsCurrent = true
	repository.activations[activation.ID] = activation
	return activation, nil
}

// Current returns the current activation.
func (repository *fakeActivationRepository) Current(
	_ context.Context,
	environment domain.ActivationEnvironment,
) (domain.ThemeActivation, error) {
	for _, activation := range repository.activations {
		if activation.Environment == environment && activation.IsCurrent {
			return activation, nil
		}
	}
	return domain.ThemeActivation{}, port.ErrNotFound
}

// FindByID returns one activation.
func (repository *fakeActivationRepository) FindByID(
	_ context.Context,
	id uuid.UUID,
) (domain.ThemeActivation, error) {
	activation, ok := repository.activations[id]
	if !ok {
		return domain.ThemeActivation{}, port.ErrNotFound
	}
	return activation, nil
}

// ListByTheme returns activation history.
func (repository *fakeActivationRepository) ListByTheme(
	_ context.Context,
	themeID uuid.UUID,
) ([]domain.ThemeActivation, error) {
	items := make([]domain.ThemeActivation, 0)
	for _, activation := range repository.activations {
		if activation.ThemeID == themeID {
			items = append(items, activation)
		}
	}
	return items, nil
}

// fakePermissions returns a configured permission error.
type fakePermissions struct {
	err error
}

// CanActivate returns the configured permission result.
func (permissions fakePermissions) CanActivate(
	context.Context,
	*uuid.UUID,
	domain.ActivationEnvironment,
) error {
	return permissions.err
}

// fakeEvents records activation events.
type fakeEvents struct {
	calls int
}

// ThemeActivated records one activation.
func (events *fakeEvents) ThemeActivated(context.Context, domain.ThemeActivation) error {
	events.calls++
	return nil
}
