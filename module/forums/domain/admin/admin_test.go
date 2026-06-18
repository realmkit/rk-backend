package admin

import (
	"testing"

	"github.com/google/uuid"
)

// TestForumPermissionGrantNormalizeAndValidate covers subject-specific grant rules.
func TestForumPermissionGrantNormalizeAndValidate(t *testing.T) {
	publicGrant := ForumPermissionGrant{
		SubjectType: " public ",
	}.Normalize()
	if publicGrant.SubjectID != PublicPermissionSubjectID() {
		t.Fatalf("expected public reserved subject id, got %s", publicGrant.SubjectID)
	}
	if violations := publicGrant.Validate("viewers[0]"); len(violations) != 0 {
		t.Fatalf("expected public grant to validate: %#v", violations)
	}

	authenticatedGrant := ForumPermissionGrant{
		SubjectType: PermissionSubjectAuthenticated,
	}.Normalize()
	if authenticatedGrant.SubjectID != AuthenticatedPermissionSubjectID() {
		t.Fatalf("expected authenticated reserved subject id, got %s", authenticatedGrant.SubjectID)
	}

	groupGrant := ForumPermissionGrant{
		SubjectType: PermissionSubjectGroup,
		SubjectID:   uuid.New(),
	}.Normalize()
	if violations := groupGrant.Validate("thread_managers[0]"); len(violations) != 0 {
		t.Fatalf("expected group grant to validate: %#v", violations)
	}

	invalidCases := []ForumPermissionGrant{
		{
			SubjectType: PermissionSubjectPublic,
			SubjectID:   uuid.New(),
		},
		{
			SubjectType: PermissionSubjectAuthenticated,
			SubjectID:   uuid.New(),
		},
		{
			SubjectType: PermissionSubjectUser,
		},
		{
			SubjectType: PermissionSubjectGroup,
		},
		{
			SubjectType: "team",
		},
	}

	for _, grant := range invalidCases {
		if violations := grant.Validate("grant"); len(violations) == 0 {
			t.Fatalf("expected invalid grant to fail: %#v", grant)
		}
	}
}

// TestForumPermissionSettingsNormalizeAndValidate covers nested grant validation.
func TestForumPermissionSettingsNormalizeAndValidate(t *testing.T) {
	settings := ForumPermissionSettings{
		ForumID: uuid.New(),
		Viewers: []ForumPermissionGrant{
			{SubjectType: PermissionSubjectPublic},
		},
		Creators: []ForumPermissionGrant{
			{SubjectType: PermissionSubjectAuthenticated},
		},
		Administrators: []ForumPermissionGrant{
			{SubjectType: PermissionSubjectGroup, SubjectID: uuid.New()},
		},
	}.Normalize()

	if err := settings.Validate(); err != nil {
		t.Fatalf("expected permission settings to validate: %v", err)
	}
	if settings.Viewers[0].SubjectID != PublicPermissionSubjectID() {
		t.Fatalf("expected normalized public viewer id")
	}
	settings.ForumID = uuid.Nil
	settings.Likers = []ForumPermissionGrant{{SubjectType: PermissionSubjectUser}}
	if err := settings.Validate(); err == nil {
		t.Fatalf("expected invalid settings to fail")
	}
}

// TestForumPermissionSettingsRejectsPublicNonViewGrants covers anonymous write denial.
func TestForumPermissionSettingsRejectsPublicNonViewGrants(t *testing.T) {
	settings := ForumPermissionSettings{
		ForumID:  uuid.New(),
		Creators: []ForumPermissionGrant{{SubjectType: PermissionSubjectPublic}},
	}.Normalize()

	if err := settings.Validate(); err == nil {
		t.Fatalf("expected public creator grant to fail")
	}
}

// TestForumPermissionSimulationRequestNormalizeAndValidate covers defaults and validation.
func TestForumPermissionSimulationRequestNormalizeAndValidate(t *testing.T) {
	forumID := uuid.New()
	request := ForumPermissionSimulationRequest{
		Permission: " forums.view ",
	}.Normalize(forumID)

	if request.Permission != "forums.view" {
		t.Fatalf("expected trimmed permission, got %q", request.Permission)
	}
	if request.ObjectType != "forum" {
		t.Fatalf("expected default forum object type, got %q", request.ObjectType)
	}
	if request.ObjectID != forumID {
		t.Fatalf("expected default object id to be forum id")
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("expected request to validate: %v", err)
	}

	if err := (ForumPermissionSimulationRequest{}).Validate(); err == nil {
		t.Fatalf("expected empty simulation request to fail")
	}
}
