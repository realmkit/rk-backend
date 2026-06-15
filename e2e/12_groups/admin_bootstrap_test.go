package groups_e2e

import (
	"context"
	"testing"

	"github.com/google/uuid"
	groupsapplication "github.com/realmkit/rk-backend/module/groups/application"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
)

// bootstrapGroupsAdmin grants the fixture actor all group administration rights.
func bootstrapGroupsAdmin(t *testing.T, service groupsapplication.Service, userID uuid.UUID) {
	t.Helper()
	admin, err := service.Create(context.Background(), groupsport.CreateGroupCommand{
		Group: groupsdomain.Group{
			Key:    "zz_e2e_admin",
			Name:   "E2E Admin",
			Color:  "#3366ff",
			Weight: 1000,
			Status: groupsdomain.GroupStatusSystem,
		},
	})
	if err != nil {
		t.Fatalf("create admin group error = %v", err)
	}
	if _, err := service.Assign(context.Background(), groupsport.AssignMembershipCommand{
		Membership: groupsdomain.Membership{
			GroupID: admin.ID,
			UserID:  userID,
			Status:  groupsdomain.MembershipStatusActive,
		},
	}); err != nil {
		t.Fatalf("assign admin membership error = %v", err)
	}
	for _, action := range groupAdminActions() {
		if _, err := service.CreatePermissionGrant(context.Background(), groupsport.CreatePermissionGrantCommand{
			GroupID: admin.ID,
			Grant: groupsdomain.PermissionGrant{
				Action:    action,
				ScopeType: groupsdomain.ObjectGroup,
				ScopeID:   groupsdomain.AllScopeID(),
			},
		}); err != nil {
			t.Fatalf("create admin grant %s error = %v", action, err)
		}
	}
	decision, err := service.Check(context.Background(), groupsport.CheckRequest{
		ActorUserID: userID,
		Action:      groupsdomain.PermissionGroupsCreate,
		ScopeType:   groupsdomain.ObjectGroup,
		ScopeID:     uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	})
	if err != nil {
		t.Fatalf("admin permission check error = %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("admin permission check denied: %+v", decision)
	}
}

// groupAdminActions returns group actions required by e2e setup.
func groupAdminActions() []groupsdomain.Action {
	return []groupsdomain.Action{
		groupsdomain.PermissionGroupsCreate,
		groupsdomain.PermissionGroupsRead,
		groupsdomain.PermissionGroupsUpdate,
		groupsdomain.PermissionGroupsDelete,
		groupsdomain.PermissionGroupsAssignMember,
		groupsdomain.PermissionGroupsReadMembers,
		groupsdomain.PermissionGroupsManagePermissions,
	}
}
