package httpguard

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/pkg/api/problem"
)

// TestTargetsBuildExpectedScopes verifies all-resource and object targets.
func TestTargetsBuildExpectedScopes(t *testing.T) {
	all := All(domain.PermissionGroupsRead, domain.ObjectGroup)
	if all.ScopeID != allGuardScopeID || all.Action != domain.PermissionGroupsRead {
		t.Fatalf("All() = %#v", all)
	}
	scopeID := uuid.New()
	object := Object(domain.PermissionGroupsManagePermissions, domain.ObjectGroup, scopeID)
	if object.ScopeID != scopeID || object.ScopeType != domain.ObjectGroup {
		t.Fatalf("Object() = %#v", object)
	}
}

// TestDeniedReturnsProblemError verifies permission denial shape.
func TestDeniedReturnsProblemError(t *testing.T) {
	err, ok := Denied().(problem.Error)
	if !ok {
		t.Fatalf("Denied() type = %T, want problem.Error", Denied())
	}
	if err.Problem.Status != 403 || err.Problem.Title == "" {
		t.Fatalf("Denied() = %#v", err.Problem)
	}
}
