package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestGroupValidateAcceptsValidGroup verifies valid group data.
func TestGroupValidateAcceptsValidGroup(t *testing.T) {
	if err := validGroup("admin", 100).Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestGroupValidateRejectsInvalidValues verifies group validation failures.
func TestGroupValidateRejectsInvalidValues(t *testing.T) {
	err := Group{Key: "No", Color: "red", Status: "missing", Weight: -1}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 5 {
		t.Fatalf("Violations = %d, want at least 5", len(validation.Violations))
	}
}

// TestMembershipActiveAtHonorsStatusAndExpiry verifies active membership rules.
func TestMembershipActiveAtHonorsStatusAndExpiry(t *testing.T) {
	now := time.Now().UTC()
	expiredAt := now.Add(-time.Minute)
	startsAt := now.Add(time.Minute)
	if (Membership{Status: MembershipStatusActive, ExpiresAt: &expiredAt}).ActiveAt(now) {
		t.Fatalf("ActiveAt() = true, want false for expired membership")
	}
	if (Membership{Status: MembershipStatusActive, StartsAt: &startsAt}).ActiveAt(now) {
		t.Fatalf("ActiveAt() = true, want false for future membership")
	}
	if (Membership{Status: MembershipStatusDisabled}).ActiveAt(now) {
		t.Fatalf("ActiveAt() = true, want false for disabled membership")
	}
	if !(Membership{Status: MembershipStatusActive}).ActiveAt(now) {
		t.Fatalf("ActiveAt() = false, want true")
	}
}

// TestMembershipValidateRejectsInvalidValues verifies membership validation.
func TestMembershipValidateRejectsInvalidValues(t *testing.T) {
	startsAt := time.Now().UTC()
	expiresAt := startsAt.Add(-time.Minute)
	err := Membership{Status: "unknown", StartsAt: &startsAt, ExpiresAt: &expiresAt, AssignedReason: strings.Repeat("x", 501)}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 5 {
		t.Fatalf("Violations = %d, want at least 5", len(validation.Violations))
	}
}

// TestDisplayGroupUsesWeightThenCreatedAtThenKey verifies display tie-breakers.
func TestDisplayGroupUsesWeightThenCreatedAtThenKey(t *testing.T) {
	now := time.Now().UTC()
	member := validGroup("member", 10)
	vip := validGroup("vip", 50)
	groups := []Group{member, vip}
	memberships := []Membership{
		{GroupID: member.ID, Status: MembershipStatusActive, CreatedAt: now.Add(-time.Hour)},
		{GroupID: vip.ID, Status: MembershipStatusActive, CreatedAt: now},
	}

	got, ok := DisplayGroup(groups, memberships, now)
	if !ok || got.ID != vip.ID {
		t.Fatalf("DisplayGroup() = (%+v, %v), want vip", got, ok)
	}
}

// TestDisplayGroupTieBreaksByCreatedAtThenKey verifies deterministic equal-weight selection.
func TestDisplayGroupTieBreaksByCreatedAtThenKey(t *testing.T) {
	now := time.Now().UTC()
	alpha := validGroup("alpha", 10)
	beta := validGroup("beta", 10)
	groups := []Group{beta, alpha}
	memberships := []Membership{
		{GroupID: beta.ID, Status: MembershipStatusActive, CreatedAt: now},
		{GroupID: alpha.ID, Status: MembershipStatusActive, CreatedAt: now},
	}

	got, ok := DisplayGroup(groups, memberships, now)
	if !ok || got.ID != alpha.ID {
		t.Fatalf("DisplayGroup() = (%+v, %v), want alpha by key tie-break", got, ok)
	}

	memberships[0].CreatedAt = now.Add(-time.Hour)
	got, ok = DisplayGroup(groups, memberships, now)
	if !ok || got.ID != beta.ID {
		t.Fatalf("DisplayGroup() = (%+v, %v), want beta by older membership", got, ok)
	}
}

// TestPermissionGrantValidateRejectsMissingID verifies grant validation.
func TestPermissionGrantValidateRejectsMissingID(t *testing.T) {
	err := PermissionGrant{Action: "groups.update", ScopeType: ObjectGroup}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) != 1 {
		t.Fatalf("Violations = %d, want 1", len(validation.Violations))
	}
}

// TestPermissionGrantValidateAcceptsAllScopeGrant verifies wildcard scopes are valid.
func TestPermissionGrantValidateAcceptsAllScopeGrant(t *testing.T) {
	grant := PermissionGrant{
		ID:        uuid.New(),
		Action:    "groups.manage_permissions",
		ScopeType: ObjectGroup,
		ScopeID:   AllScopeID(),
	}
	if err := grant.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !grant.AppliesToAllScopes() {
		t.Fatalf("AppliesToAllScopes() = false, want true")
	}
}

// TestPermissionActionValidateRejectsInvalidAction verifies action validation.
func TestPermissionActionValidateRejectsInvalidAction(t *testing.T) {
	err := PermissionAction{ID: uuid.New(), Action: "Posts Update", ScopeType: "post", Area: "posts", Label: "Update", WarningLevel: WarningLevelNormal, Enabled: true, Version: 1}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) == 0 {
		t.Fatalf("Violations = 0, want duration violation")
	}
}

// TestPolicyConditionValidationCoversSupportedShapes verifies condition branch validation.
func TestPolicyConditionValidationCoversSupportedShapes(t *testing.T) {
	valid := []PolicyCondition{
		{Type: ConditionEquals, Field: "thread.status", Value: "open"},
		{Type: ConditionIn, Field: "ticket.status", Values: []string{"open"}},
		{Type: ConditionFieldEqualsActor, Field: "author_user_id"},
		{Type: ConditionOlderThan, Field: "created_at", Duration: "24h"},
	}
	for index, condition := range valid {
		if violations := condition.Validate("conditions[" + string(rune('0'+index)) + "]"); len(violations) != 0 {
			t.Fatalf("condition %d violations = %+v, want none", index, violations)
		}
	}
	if violations := (PolicyCondition{Type: ConditionIn, Field: "status"}).Validate("condition"); len(violations) != 1 {
		t.Fatalf("ConditionIn without values violations = %+v, want one", violations)
	}
}

// TestPermissionActionValidateAcceptsDottedAction verifies permission names.
func TestPermissionActionValidateAcceptsDottedAction(t *testing.T) {
	action := PermissionAction{
		ID:           uuid.New(),
		Action:       "posts.update",
		Area:         "posts",
		ScopeType:    "post",
		Label:        "Update posts",
		WarningLevel: WarningLevelDangerous,
		Enabled:      true,
		Version:      1,
	}
	if err := action.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// validGroup returns a valid group.
func validGroup(key Key, weight int) Group {
	return Group{ID: uuid.New(), Key: key, Name: string(key), Color: "#ff00aa", Weight: weight, Status: GroupStatusActive, Version: 1}
}
