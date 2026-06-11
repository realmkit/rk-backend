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

// TestRelationTupleValidateRejectsMissingIdentifiers verifies tuple validation.
func TestRelationTupleValidateRejectsMissingIdentifiers(t *testing.T) {
	err := RelationTuple{ObjectType: ObjectGroup, Relation: RelationMember, SubjectType: SubjectUser}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) != 2 {
		t.Fatalf("Violations = %d, want 2", len(validation.Violations))
	}
}

// TestRelationTupleValidateAcceptsPublicSubject verifies public tuple validation.
func TestRelationTupleValidateAcceptsPublicSubject(t *testing.T) {
	tuple := RelationTuple{
		ObjectType:  ObjectForum,
		ObjectID:    uuid.New(),
		Relation:    RelationViewer,
		SubjectType: SubjectPublic,
		SubjectID:   PublicSubjectID(),
	}
	if err := tuple.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestRelationTupleValidateRejectsWrongReservedSubjectID verifies reserved subject IDs.
func TestRelationTupleValidateRejectsWrongReservedSubjectID(t *testing.T) {
	tuple := RelationTuple{
		ObjectType:  ObjectForum,
		ObjectID:    uuid.New(),
		Relation:    RelationViewer,
		SubjectType: SubjectAuthenticated,
		SubjectID:   uuid.New(),
	}
	err := tuple.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) == 0 {
		t.Fatalf("Violations = 0, want reserved subject violation")
	}
}

// TestPermissionRuleValidateRejectsInvalidConditions verifies condition validation.
func TestPermissionRuleValidateRejectsInvalidConditions(t *testing.T) {
	err := PermissionRule{
		ID:         uuid.New(),
		Permission: "posts.update",
		ObjectType: "post",
		Relation:   RelationEditor,
		Conditions: []PolicyCondition{{Type: ConditionWithinDuration, Field: "post.created_at", Duration: "soon"}},
	}.Validate()
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

// TestPermissionDefinitionValidateAcceptsDottedPermission verifies permission names.
func TestPermissionDefinitionValidateAcceptsDottedPermission(t *testing.T) {
	definition := PermissionDefinition{ID: uuid.New(), Permission: "posts.update", ObjectType: "post", Enabled: true, Version: 1}
	if err := definition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// validGroup returns a valid group.
func validGroup(key Key, weight int) Group {
	return Group{ID: uuid.New(), Key: key, Name: string(key), Color: "#ff00aa", Weight: weight, Status: GroupStatusActive, Version: 1}
}
