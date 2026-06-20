package routedata

import (
	"context"
	"time"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// NewService creates a route-data service.
func NewService(visibility VisibilityChecker, clock Clock) Service {
	if clock == nil {
		clock = time.Now
	}
	return Service{visibility: visibility, clock: clock}
}

// Resolve returns route data after visibility checks.
func (service Service) Resolve(ctx context.Context, request Request) (domain.RouteDataEnvelope, error) {
	contract, err := ContractFor(request.Route)
	if err != nil {
		return domain.RouteDataEnvelope{}, err
	}
	if err := validateParams(request, contract); err != nil {
		return domain.RouteDataEnvelope{}, err
	}
	if service.visibility != nil {
		if err := service.visibility.CanViewRoute(ctx, request, contract); err != nil {
			return domain.RouteDataEnvelope{}, err
		}
	}
	if request.Now.IsZero() {
		request.Now = service.clock().UTC()
	}
	return envelope(request, contract), nil
}

// PreviewPersonas returns supported preview persona kinds.
func PreviewPersonas() []domain.PreviewPersonaKind {
	return []domain.PreviewPersonaKind{
		domain.PersonaAnonymous,
		domain.PersonaSignedIn,
		domain.PersonaGroup,
		domain.PersonaModerator,
		domain.PersonaUser,
	}
}

// validateParams verifies required route parameters.
func validateParams(request Request, contract Contract) error {
	for _, key := range contract.RequiredParams {
		if request.Params[key] == "" {
			return port.ErrInvalidState
		}
	}
	return nil
}
