package application

import (
	"context"
	"testing"

	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/principal"
)

// benchmarkPrincipal stores the provision benchmark result.
var benchmarkPrincipal principal.Principal

// BenchmarkProvisionExistingIdentity measures the steady-state identity provisioning path.
func BenchmarkProvisionExistingIdentity(b *testing.B) {
	service, _, _, _ := newTestService()
	external := testIdentity()
	token := auth.Token{Identity: external, Audience: []string{"api"}, Scopes: []string{"openid", "profile"}}
	ctx := context.Background()
	if _, err := service.Provision(ctx, external, token); err != nil {
		b.Fatalf("initial Provision() error = %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		principal, err := service.Provision(ctx, external, token)
		if err != nil {
			b.Fatalf("Provision() error = %v", err)
		}
		benchmarkPrincipal = principal
	}
}
