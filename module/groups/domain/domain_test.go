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
	if (Membership{Status: MembershipStatusActive, ExpiresAt: &expiredAt}).ActiveAt(now) {
		t.Fatalf("ActiveAt() = true, want false for expired membership")
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
