package application

import (
	"testing"

	"github.com/realmkit/rk-backend/module/groups/domain"
)

// TestThemePermissionActionsAreRegistered verifies theme actions are exposed to permission checks.
func TestThemePermissionActionsAreRegistered(t *testing.T) {
	for _, action := range []domain.Action{
		"themes.view",
		"themes.import",
		"themes.edit",
		"themes.validate",
		"themes.publish",
		"themes.rollback",
		"themes.delete",
		"themes.preview",
		"themes.activate",
	} {
		permission, err := permissionAction(action)
		if err != nil {
			t.Fatalf("permissionAction(%q) error = %v", action, err)
		}
		if permission.ScopeType != domain.ObjectTheme {
			t.Fatalf("permissionAction(%q).ScopeType = %q, want %q", action, permission.ScopeType, domain.ObjectTheme)
		}
	}
}
