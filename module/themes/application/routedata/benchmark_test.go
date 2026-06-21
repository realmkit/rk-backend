package routedata

import (
	"context"
	"testing"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// benchmarkRouteEnvelope stores route data benchmark output.
var benchmarkRouteEnvelope domain.RouteDataEnvelope

// BenchmarkResolve measures route contract lookup, parameter validation, visibility, and envelope assembly.
func BenchmarkResolve(b *testing.B) {
	service := NewService(fakeVisibility{}, nil)
	request := Request{
		Route:  domain.RouteThreadsShow,
		Locale: "en",
		Path:   "/threads/welcome",
		Params: map[string]string{"thread_slug": "welcome"},
		Theme:  testThemeContext(),
		Viewer: ViewerContext{
			PersonaKind:   domain.PersonaModerator,
			PersonaSource: domain.PersonaSourceSynthetic,
			IsPreview:     true,
		},
	}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		envelope, err := service.Resolve(ctx, request)
		if err != nil {
			b.Fatalf("Resolve() error = %v", err)
		}
		benchmarkRouteEnvelope = envelope
	}
}
