package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

// ErrInvalid reports invalid group domain data.
var ErrInvalid = errors.New("invalid group")

// Violation describes one validation failure.
type Violation struct {
	// Field is the failing field path.
	Field string `json:"field"`

	// Message explains the validation failure.
	Message string `json:"message"`
}

// ValidationError contains validation failures.
type ValidationError struct {
	// Violations contains all validation failures.
	Violations []Violation `json:"violations"`
}

// Error returns the validation error message.
func (err ValidationError) Error() string {
	return ErrInvalid.Error()
}

// Is reports whether target is ErrInvalid.
func (err ValidationError) Is(target error) bool {
	return errors.Is(target, ErrInvalid)
}

// NewValidationError returns a validation error when violations exist.
func NewValidationError(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return ValidationError{Violations: violations}
}

// AppendViolation appends one validation failure.
func AppendViolation(violations []Violation, field string, message string) []Violation {
	return append(violations, Violation{Field: field, Message: message})
}

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
	statuses := []MembershipStatus{
		MembershipStatusActive,
		MembershipStatusDisabled,
		MembershipStatusExpired,
		MembershipStatusRevoked,
	}
	if slices.Contains(statuses, status) {
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

// ValidatePermission validates a dotted permission name.
func ValidatePermission(field string, value Permission) []Violation {
	permission := strings.TrimSpace(string(value))
	if permission == "" || len(permission) > 120 {
		return []Violation{{Field: field, Message: "must be a dotted permission between 1 and 120 characters"}}
	}
	parts := strings.Split(permission, ".")
	for _, part := range parts {
		if !keyPattern.MatchString(part) {
			return []Violation{{Field: field, Message: "must use lower snake case segments separated by dots"}}
		}
	}
	return nil
}

// ValidateConditionType validates condition type.
func ValidateConditionType(field string, conditionType ConditionType) []Violation {
	types := []ConditionType{
		ConditionEquals,
		ConditionIn,
		ConditionFieldEqualsActor,
		ConditionFieldNotEqualsActor,
		ConditionIsUnset,
		ConditionAssignedToActor,
		ConditionWithinDuration,
		ConditionOlderThan,
	}
	if slices.Contains(types, conditionType) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// requireField validates a condition field name.
func requireField(field string, value string) []Violation {
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	return nil
}

// requireValue validates a condition scalar value.
func requireValue(field string, value string) []Violation {
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	return nil
}

// itoa formats a small integer without importing strconv into callers.
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[index:])
}
