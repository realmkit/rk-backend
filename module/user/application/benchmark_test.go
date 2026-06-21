package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/principal"
)

// benchmarkPrincipal stores the provision benchmark result.
var benchmarkPrincipal principal.Principal

// benchmarkUserSummaries stores the user summary benchmark result.
var benchmarkUserSummaries map[uuid.UUID]port.UserSummary

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

// BenchmarkFindSummariesByIDs measures summary map assembly for cross-module user lookups.
func BenchmarkFindSummariesByIDs(b *testing.B) {
	service, users, _, _ := newTestService()
	ids := make([]uuid.UUID, 128)
	for index := range ids {
		id := uuid.New()
		ids[index] = id
		users.items[id] = domain.User{
			ID:          id,
			Status:      domain.StatusActive,
			FirstSeenAt: time.Unix(100, 0).UTC(),
			Version:     1,
		}
	}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		summaries, err := service.FindSummariesByIDs(ctx, ids)
		if err != nil {
			b.Fatalf("FindSummariesByIDs() error = %v", err)
		}
		benchmarkUserSummaries = summaries
	}
}
