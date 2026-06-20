package routedata

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// TestContractsCoverEveryRouteKind verifies route-data contracts stay complete.
func TestContractsCoverEveryRouteKind(t *testing.T) {
	contracts := map[domain.RouteKind]Contract{}
	for _, contract := range Contracts() {
		contracts[contract.Route] = contract
	}
	for _, route := range domain.RouteKinds() {
		if _, ok := contracts[route]; !ok {
			t.Fatalf("route %q missing contract", route)
		}
	}
}

// TestResolveBuildsEnvelopeForPreviewPersonas verifies common route data for persona lenses.
func TestResolveBuildsEnvelopeForPreviewPersonas(t *testing.T) {
	service := NewService(fakeVisibility{}, nil)
	for _, persona := range PreviewPersonas() {
		envelope, err := service.Resolve(context.Background(), Request{
			Route:  domain.RouteThreadsShow,
			Locale: "en",
			Path:   "/threads/welcome",
			Params: map[string]string{"thread_slug": "welcome"},
			Theme:  testThemeContext(),
			Viewer: ViewerContext{PersonaKind: persona, PersonaSource: domain.PersonaSourceSynthetic, IsPreview: true},
		})
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", persona, err)
		}
		if envelope.Viewer["persona"] != persona || envelope.Metadata["rich_text"] == nil {
			t.Fatalf("envelope = %+v, want persona and rich text metadata", envelope)
		}
	}
}

// TestResolveRejectsMissingParamsAndVisibilityDenials verifies route guards.
func TestResolveRejectsMissingParamsAndVisibilityDenials(t *testing.T) {
	service := NewService(nil, nil)
	_, err := service.Resolve(context.Background(), Request{Route: domain.RouteForumsShow, Theme: testThemeContext()})
	if !errors.Is(err, port.ErrInvalidState) {
		t.Fatalf("missing param error = %v, want invalid state", err)
	}
	denied := NewService(fakeVisibility{err: port.ErrPermissionDenied}, nil)
	_, err = denied.Resolve(context.Background(), Request{Route: domain.RouteHome, Theme: testThemeContext()})
	if !errors.Is(err, port.ErrPermissionDenied) {
		t.Fatalf("visibility error = %v, want permission denied", err)
	}
}

// fakeVisibility records route visibility checks.
type fakeVisibility struct {
	err error
}

// CanViewRoute returns the configured error.
func (visibility fakeVisibility) CanViewRoute(context.Context, Request, Contract) error {
	return visibility.err
}

// testThemeContext returns a route-data theme context.
func testThemeContext() ThemeContext {
	return ThemeContext{
		ThemeID: uuid.New(), VersionID: uuid.New(), ActivationID: uuid.New(),
		Environment: domain.EnvironmentPublic, SettingsData: map[string]any{"brand": "RealmKit"},
		IntegritySHA256: "version-sha",
	}
}
