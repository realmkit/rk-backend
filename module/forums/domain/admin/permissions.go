package admin

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ForumPermissionGrant grants one forum relation to one subject.
type ForumPermissionGrant struct {
	// SubjectType is public, authenticated, user, or group.
	SubjectType PermissionSubjectType `json:"subject_type"`

	// SubjectID is the concrete subject identifier.
	SubjectID uuid.UUID `json:"subject_id"`

	// SubjectRelation is usually member for group subjects.
	SubjectRelation string `json:"subject_relation,omitempty"`
}

// Normalize returns a normalized permission grant.
func (grant ForumPermissionGrant) Normalize() ForumPermissionGrant {
	grant.SubjectType = PermissionSubjectType(strings.TrimSpace(string(grant.SubjectType)))
	grant.SubjectRelation = strings.TrimSpace(grant.SubjectRelation)
	switch grant.SubjectType {
	case PermissionSubjectPublic:
		grant.SubjectID = PublicPermissionSubjectID()
	case PermissionSubjectAuthenticated:
		grant.SubjectID = AuthenticatedPermissionSubjectID()
	case PermissionSubjectGroup:
		if grant.SubjectRelation == "" {
			grant.SubjectRelation = "member"
		}
	}
	return grant
}

// Validate validates a forum permission grant.
func (grant ForumPermissionGrant) Validate(field string) []Violation {
	var violations []Violation
	violations = append(
		violations,
		ValidatePermissionSubjectType(field+".subject_type", grant.SubjectType)...,
	)
	switch grant.SubjectType {
	case PermissionSubjectPublic:
		violations = grant.validatePublic(field, violations)
	case PermissionSubjectAuthenticated:
		violations = grant.validateAuthenticated(field, violations)
	case PermissionSubjectUser:
		violations = grant.validateUser(field, violations)
	case PermissionSubjectGroup:
		violations = grant.validateGroup(field, violations)
	}
	return violations
}

// ForumPermissionSettings contains forum relation grants.
type ForumPermissionSettings struct {
	// ForumID is the configured forum.
	ForumID uuid.UUID `json:"forum_id"`

	// Viewers can view the forum.
	Viewers []ForumPermissionGrant `json:"viewers"`

	// Creators can create threads.
	Creators []ForumPermissionGrant `json:"creators"`

	// Replyers can reply to threads.
	Replyers []ForumPermissionGrant `json:"replyers"`

	// Likers can like posts.
	Likers []ForumPermissionGrant `json:"likers"`

	// Moderators can moderate threads, posts, and sticky state.
	Moderators []ForumPermissionGrant `json:"moderators"`

	// Managers can manage forum settings and permissions.
	Managers []ForumPermissionGrant `json:"managers"`
}

// Normalize returns normalized permission settings.
func (settings ForumPermissionSettings) Normalize() ForumPermissionSettings {
	settings.Viewers = normalizePermissionGrants(settings.Viewers)
	settings.Creators = normalizePermissionGrants(settings.Creators)
	settings.Replyers = normalizePermissionGrants(settings.Replyers)
	settings.Likers = normalizePermissionGrants(settings.Likers)
	settings.Moderators = normalizePermissionGrants(settings.Moderators)
	settings.Managers = normalizePermissionGrants(settings.Managers)
	return settings
}

// Validate validates permission settings.
func (settings ForumPermissionSettings) Validate() error {
	var violations []Violation
	if settings.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	violations = append(violations, validatePermissionGrants("viewers", settings.Viewers)...)
	violations = append(violations, validatePermissionGrants("creators", settings.Creators)...)
	violations = append(violations, validatePermissionGrants("replyers", settings.Replyers)...)
	violations = append(violations, validatePermissionGrants("likers", settings.Likers)...)
	violations = append(violations, validatePermissionGrants("moderators", settings.Moderators)...)
	violations = append(violations, validatePermissionGrants("managers", settings.Managers)...)
	return NewValidationError(violations)
}

// PublicPermissionSubjectID returns the public grant subject identifier.
func PublicPermissionSubjectID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// AuthenticatedPermissionSubjectID returns the authenticated grant subject identifier.
func AuthenticatedPermissionSubjectID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000002")
}

func normalizePermissionGrants(grants []ForumPermissionGrant) []ForumPermissionGrant {
	normalized := make([]ForumPermissionGrant, 0, len(grants))
	for _, grant := range grants {
		normalized = append(normalized, grant.Normalize())
	}
	return normalized
}

func validatePermissionGrants(field string, grants []ForumPermissionGrant) []Violation {
	var violations []Violation
	for index, grant := range grants {
		itemField := field + "[" + strconv.Itoa(index) + "]"
		violations = append(violations, grant.Validate(itemField)...)
	}
	return violations
}

func (grant ForumPermissionGrant) validatePublic(
	field string,
	violations []Violation,
) []Violation {
	if grant.SubjectID != PublicPermissionSubjectID() {
		violations = AppendViolation(violations, field+".subject_id", "must use the public reserved identifier")
	}
	if grant.SubjectRelation != "" {
		violations = AppendViolation(violations, field+".subject_relation", "must be empty for public grants")
	}
	return violations
}

func (grant ForumPermissionGrant) validateAuthenticated(
	field string,
	violations []Violation,
) []Violation {
	if grant.SubjectID != AuthenticatedPermissionSubjectID() {
		violations = AppendViolation(violations, field+".subject_id", "must use the authenticated reserved identifier")
	}
	if grant.SubjectRelation != "" {
		violations = AppendViolation(violations, field+".subject_relation", "must be empty for authenticated grants")
	}
	return violations
}

func (grant ForumPermissionGrant) validateUser(
	field string,
	violations []Violation,
) []Violation {
	if grant.SubjectID == uuid.Nil {
		violations = AppendViolation(violations, field+".subject_id", "is required")
	}
	if grant.SubjectRelation != "" {
		violations = AppendViolation(violations, field+".subject_relation", "must be empty for user grants")
	}
	return violations
}

func (grant ForumPermissionGrant) validateGroup(
	field string,
	violations []Violation,
) []Violation {
	if grant.SubjectID == uuid.Nil {
		violations = AppendViolation(violations, field+".subject_id", "is required")
	}
	if grant.SubjectRelation != "member" {
		violations = AppendViolation(violations, field+".subject_relation", "must be member for group grants")
	}
	return violations
}
