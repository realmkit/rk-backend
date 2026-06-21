package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
)

// benchmarkDecision stores the permission benchmark decision.
var benchmarkDecision port.Decision

// benchmarkPermissionActions stores the permission catalog benchmark result.
var benchmarkPermissionActions []domain.PermissionAction

// BenchmarkCheckPermissionGrant measures the application permission-check path over in-memory group grants.
func BenchmarkCheckPermissionGrant(b *testing.B) {
	service, groups, memberships, permissions := newTestService()
	group := testGroup("moderator", domain.GroupStatusActive)
	actorID := uuid.New()
	scopeID := uuid.New()
	grantID := uuid.New()
	groups.items[group.ID] = group
	memberships.items[membershipKey(group.ID, actorID)] = domain.Membership{
		ID:      uuid.New(),
		GroupID: group.ID,
		UserID:  actorID,
		Status:  domain.MembershipStatusActive,
		Version: 1,
	}
	permissions.grants[grantID] = domain.PermissionGrant{
		ID:        grantID,
		Action:    "assets.view",
		ScopeType: domain.ObjectAsset,
		ScopeID:   domain.AllScopeID(),
	}
	permissions.assignments[grantID] = group.ID
	request := port.CheckRequest{
		ActorUserID: actorID,
		Action:      "assets.view",
		ScopeType:   domain.ObjectAsset,
		ScopeID:     scopeID,
	}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		decision, err := service.Check(ctx, request)
		if err != nil {
			b.Fatalf("Check() error = %v", err)
		}
		benchmarkDecision = decision
	}
}

// BenchmarkListPermissionActions measures permission catalog materialization and sorting.
func BenchmarkListPermissionActions(b *testing.B) {
	service, _, _, _ := newTestService()
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		actions, err := service.ListPermissionActions(ctx)
		if err != nil {
			b.Fatalf("ListPermissionActions() error = %v", err)
		}
		benchmarkPermissionActions = actions
	}
}
