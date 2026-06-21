package shared

import (
	"encoding/json"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// keyPattern matches stable lower snake identifiers.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// slugPattern matches URL slugs.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,118}[a-z0-9]$`)

// ValidateKey validates key.
func ValidateKey(field string, key Key) []Violation {
	if keyPattern.MatchString(strings.TrimSpace(string(key))) {
		return nil
	}
	return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
}

// ValidateSlug validates slug.
func ValidateSlug(field string, slug Slug) []Violation {
	if slugPattern.MatchString(strings.TrimSpace(string(slug))) {
		return nil
	}
	return []Violation{{Field: field, Message: "must be lower kebab case and between 3 and 120 characters"}}
}

// ValidateName validates display names.
func ValidateName(field string, value string) []Violation {
	length := len(strings.TrimSpace(value))
	if length >= 1 && length <= 120 {
		return nil
	}
	return []Violation{{Field: field, Message: "must be between 1 and 120 characters"}}
}

// ValidateDescription validates descriptions.
func ValidateDescription(field string, value string) []Violation {
	if len(strings.TrimSpace(value)) <= 1000 {
		return nil
	}
	return []Violation{{Field: field, Message: "must be at most 1000 characters"}}
}

// ValidateDisplayOrder validates display order.
func ValidateDisplayOrder(field string, value int) []Violation {
	if value >= 0 {
		return nil
	}
	return []Violation{{Field: field, Message: "must be zero or greater"}}
}

// ValidateCategoryStatus validates category status.
func ValidateCategoryStatus(field string, status CategoryStatus) []Violation {
	allowed := []CategoryStatus{
		CategoryStatusActive,
		CategoryStatusHidden,
		CategoryStatusArchived,
	}
	return validateEnum(field, slices.Contains(allowed, status))
}

// ValidateForumKind validates forum kind.
func ValidateForumKind(field string, kind ForumKind) []Violation {
	allowed := []ForumKind{ForumKindDiscussion, ForumKindLink, ForumKindContainer}
	return validateEnum(field, slices.Contains(allowed, kind))
}

// ValidateForumStatus validates forum status.
func ValidateForumStatus(field string, status ForumStatus) []Violation {
	allowed := []ForumStatus{ForumStatusActive, ForumStatusHidden, ForumStatusArchived}
	return validateEnum(field, slices.Contains(allowed, status))
}

// ValidateThreadVisibilityMode validates thread visibility mode.
func ValidateThreadVisibilityMode(field string, mode ThreadVisibilityMode) []Violation {
	allowed := []ThreadVisibilityMode{
		ThreadVisibilityAllThreads,
		ThreadVisibilityOwnThreads,
		ThreadVisibilityOwnOrStickyThreads,
	}
	return validateEnum(field, slices.Contains(allowed, mode))
}

// ValidateThreadStatus validates thread status.
func ValidateThreadStatus(field string, status ThreadStatus) []Violation {
	allowed := []ThreadStatus{
		ThreadStatusOpen,
		ThreadStatusClosed,
		ThreadStatusLocked,
		ThreadStatusHidden,
		ThreadStatusArchived,
		ThreadStatusDeleted,
	}
	return validateEnum(field, slices.Contains(allowed, status))
}

// ValidateStickyState validates sticky state.
func ValidateStickyState(field string, state StickyState) []Violation {
	allowed := []StickyState{
		StickyStateNormal,
		StickyStateSticky,
		StickyStateAnnouncement,
	}
	return validateEnum(field, slices.Contains(allowed, state))
}

// ValidatePostStatus validates post status.
func ValidatePostStatus(field string, status PostStatus) []Violation {
	allowed := []PostStatus{
		PostStatusVisible,
		PostStatusHidden,
		PostStatusPendingReview,
		PostStatusDeleted,
		PostStatusSystem,
	}
	return validateEnum(field, slices.Contains(allowed, status))
}

// ValidateContentFormat validates content format.
func ValidateContentFormat(field string, format ContentFormat) []Violation {
	return validateEnum(field, format == ContentFormatProseMirror)
}

// ValidateReferenceType validates reference type.
func ValidateReferenceType(field string, referenceType ReferenceType) []Violation {
	allowed := []ReferenceType{
		ReferenceReplyTo,
		ReferenceQuote,
		ReferenceMention,
		ReferenceAttachment,
		ReferenceLink,
	}
	return validateEnum(field, slices.Contains(allowed, referenceType))
}

// ValidatePermissionSubjectType validates forum permission grant subjects.
func ValidatePermissionSubjectType(
	field string,
	subjectType PermissionSubjectType,
) []Violation {
	allowed := []PermissionSubjectType{
		PermissionSubjectPublic,
		PermissionSubjectAuthenticated,
		PermissionSubjectUser,
		PermissionSubjectGroup,
	}
	return validateEnum(field, slices.Contains(allowed, subjectType))
}

// ValidateExternalURL validates optional external URL.
func ValidateExternalURL(field string, value string) []Violation {
	if strings.TrimSpace(value) == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return nil
	}
	return []Violation{{Field: field, Message: "must be an absolute URL"}}
}

// ValidateTitle validates thread titles.
func ValidateTitle(field string, value string) []Violation {
	length := len(strings.TrimSpace(value))
	if length >= 3 && length <= 160 {
		return nil
	}
	return []Violation{{Field: field, Message: "must be between 3 and 160 characters"}}
}

// ValidateContentDocument validates stored rich-content JSON.
func ValidateContentDocument(field string, document json.RawMessage) []Violation {
	if len(document) == 0 {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if len(document) > 65536 {
		return []Violation{{Field: field, Message: "must be at most 65536 bytes"}}
	}
	var payload any
	if err := json.Unmarshal(document, &payload); err != nil {
		return []Violation{{Field: field, Message: "must be valid JSON"}}
	}
	if _, ok := payload.(map[string]any); !ok {
		return []Violation{{Field: field, Message: "must be a JSON object"}}
	}
	return nil
}

// ValidateContentText validates extracted text.
func ValidateContentText(field string, value string) []Violation {
	length := len(strings.TrimSpace(value))
	if length >= 1 && length <= 200000 {
		return nil
	}
	return []Violation{{Field: field, Message: "must be between 1 and 200000 characters"}}
}

// validateEnum supports package behavior.
func validateEnum(field string, allowed bool) []Violation {
	if allowed {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}
