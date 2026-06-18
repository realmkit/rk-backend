package admin

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ForumPermissionGrant grants one forum action to one subject.
type ForumPermissionGrant struct {
	// SubjectType is public, authenticated, user, or group.
	SubjectType PermissionSubjectType `json:"subject_type"`

	// SubjectID is the concrete subject identifier.
	SubjectID uuid.UUID `json:"subject_id"`
}

// Normalize returns a normalized permission grant.
func (grant ForumPermissionGrant) Normalize() ForumPermissionGrant {
	grant.SubjectType = PermissionSubjectType(strings.TrimSpace(string(grant.SubjectType)))
	switch grant.SubjectType {
	case PermissionSubjectPublic:
		grant.SubjectID = PublicPermissionSubjectID()
	case PermissionSubjectAuthenticated:
		grant.SubjectID = AuthenticatedPermissionSubjectID()
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

	// ThreadPinners can pin and unpin threads.
	ThreadPinners []ForumPermissionGrant `json:"thread_pinners"`

	// ThreadManagers can close, open, update, or delete threads.
	ThreadManagers []ForumPermissionGrant `json:"thread_managers"`

	// PostManagers can update or delete posts.
	PostManagers []ForumPermissionGrant `json:"post_managers"`

	// LimitBypassers can bypass configured thread limits.
	LimitBypassers []ForumPermissionGrant `json:"limit_bypassers"`

	// AllThreadViewers can see all threads despite forum visibility filtering.
	AllThreadViewers []ForumPermissionGrant `json:"all_thread_viewers"`

	// Administrators receive all forum moderation permissions.
	Administrators []ForumPermissionGrant `json:"administrators"`
}

// Normalize returns normalized permission settings.
func (settings ForumPermissionSettings) Normalize() ForumPermissionSettings {
	settings.Viewers = normalizePermissionGrants(settings.Viewers)
	settings.Creators = normalizePermissionGrants(settings.Creators)
	settings.Replyers = normalizePermissionGrants(settings.Replyers)
	settings.Likers = normalizePermissionGrants(settings.Likers)
	settings.ThreadPinners = normalizePermissionGrants(settings.ThreadPinners)
	settings.ThreadManagers = normalizePermissionGrants(settings.ThreadManagers)
	settings.PostManagers = normalizePermissionGrants(settings.PostManagers)
	settings.LimitBypassers = normalizePermissionGrants(settings.LimitBypassers)
	settings.AllThreadViewers = normalizePermissionGrants(settings.AllThreadViewers)
	settings.Administrators = normalizePermissionGrants(settings.Administrators)
	return settings
}

// Validate validates permission settings.
func (settings ForumPermissionSettings) Validate() error {
	var violations []Violation
	if settings.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	violations = append(violations, validatePermissionGrants("viewers", settings.Viewers)...)
	violations = append(violations, validatePrivatePermissionGrants("creators", settings.Creators)...)
	violations = append(violations, validatePrivatePermissionGrants("replyers", settings.Replyers)...)
	violations = append(violations, validatePrivatePermissionGrants("likers", settings.Likers)...)
	violations = append(violations, validatePrivatePermissionGrants("thread_pinners", settings.ThreadPinners)...)
	violations = append(violations, validatePrivatePermissionGrants("thread_managers", settings.ThreadManagers)...)
	violations = append(violations, validatePrivatePermissionGrants("post_managers", settings.PostManagers)...)
	violations = append(violations, validatePrivatePermissionGrants("limit_bypassers", settings.LimitBypassers)...)
	violations = append(violations, validatePrivatePermissionGrants("all_thread_viewers", settings.AllThreadViewers)...)
	violations = append(violations, validatePrivatePermissionGrants("administrators", settings.Administrators)...)
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

func validatePrivatePermissionGrants(field string, grants []ForumPermissionGrant) []Violation {
	var violations []Violation
	for index, grant := range grants {
		itemField := field + "[" + strconv.Itoa(index) + "]"
		violations = append(violations, grant.Validate(itemField)...)
		if grant.SubjectType == PermissionSubjectPublic {
			violations = AppendViolation(
				violations,
				itemField+".subject_type",
				"anonymous users can only view forums",
			)
		}
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
	return violations
}

func (grant ForumPermissionGrant) validateAuthenticated(
	field string,
	violations []Violation,
) []Violation {
	if grant.SubjectID != AuthenticatedPermissionSubjectID() {
		violations = AppendViolation(violations, field+".subject_id", "must use the authenticated reserved identifier")
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
	return violations
}

func (grant ForumPermissionGrant) validateGroup(
	field string,
	violations []Violation,
) []Violation {
	if grant.SubjectID == uuid.Nil {
		violations = AppendViolation(violations, field+".subject_id", "is required")
	}
	return violations
}
