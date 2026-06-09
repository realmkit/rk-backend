package domain

import (
	"regexp"
	"slices"
	"strings"
)

// Key is a stable group key.
type Key string

// Color is a hex UI color.
type Color string

// GroupStatus is the group lifecycle state.
type GroupStatus string

// MembershipStatus is the membership lifecycle state.
type MembershipStatus string

// ObjectType identifies an authorization object type.
type ObjectType string

// Relation identifies an authorization relation.
type Relation string

// SubjectType identifies an authorization subject type.
type SubjectType string

// Permission identifies a domain action.
type Permission string

const (
	// GroupStatusActive means the group grants permissions.
	GroupStatusActive GroupStatus = "active"

	// GroupStatusDisabled means the group does not grant permissions.
	GroupStatusDisabled GroupStatus = "disabled"

	// GroupStatusSystem means the group is built in and grants permissions.
	GroupStatusSystem GroupStatus = "system"
)

const (
	// MembershipStatusActive means the membership grants permissions.
	MembershipStatusActive MembershipStatus = "active"

	// MembershipStatusDisabled means the membership is disabled.
	MembershipStatusDisabled MembershipStatus = "disabled"

	// MembershipStatusExpired means the membership is expired.
	MembershipStatusExpired MembershipStatus = "expired"

	// MembershipStatusRevoked means the membership is revoked.
	MembershipStatusRevoked MembershipStatus = "revoked"
)

const (
	// ObjectGroup is a group authorization object.
	ObjectGroup ObjectType = "group"

	// ObjectAsset is an asset authorization object.
	ObjectAsset ObjectType = "asset"

	// ObjectUser is a user authorization object.
	ObjectUser ObjectType = "user"

	// ObjectSystem is a system authorization object.
	ObjectSystem ObjectType = "system"
)

const (
	// SubjectUser is a user subject.
	SubjectUser SubjectType = "user"

	// SubjectGroup is a group subject.
	SubjectGroup SubjectType = "group"
)

const (
	// RelationMember is a member relation.
	RelationMember Relation = "member"

	// RelationViewer is a viewer relation.
	RelationViewer Relation = "viewer"

	// RelationEditor is an editor relation.
	RelationEditor Relation = "editor"

	// RelationManager is a manager relation.
	RelationManager Relation = "manager"

	// RelationOwner is an owner relation.
	RelationOwner Relation = "owner"

	// RelationSelf is a self relation.
	RelationSelf Relation = "self"
)

// keyPattern matches stable lower snake identifiers.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// colorPattern matches six-digit hex colors.
var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// ValidateKey validates key.
func ValidateKey(field string, key Key) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(string(key))) {
		return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
	}
	return nil
}

// ValidateColor validates color.
func ValidateColor(field string, color Color) []Violation {
	if !colorPattern.MatchString(strings.TrimSpace(string(color))) {
		return []Violation{{Field: field, Message: "must be a hex color"}}
	}
	return nil
}

// ValidateGroupStatus validates group status.
func ValidateGroupStatus(field string, status GroupStatus) []Violation {
	if slices.Contains([]GroupStatus{GroupStatusActive, GroupStatusDisabled, GroupStatusSystem}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateMembershipStatus validates membership status.
func ValidateMembershipStatus(field string, status MembershipStatus) []Violation {
	if slices.Contains([]MembershipStatus{MembershipStatusActive, MembershipStatusDisabled, MembershipStatusExpired, MembershipStatusRevoked}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateRelationTerm validates relation-like lower snake text.
func ValidateRelationTerm(field string, value string) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(value)) {
		return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
	}
	return nil
}
