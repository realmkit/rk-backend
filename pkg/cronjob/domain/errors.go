package domain

import "strings"

// Violation describes one cron validation failure.
type Violation struct {
	// Field is the invalid field name.
	Field string `json:"field"`

	// Message describes the validation failure.
	Message string `json:"message"`
}

// ValidationError reports invalid cron data.
type ValidationError struct {
	// Violations contains all detected validation failures.
	Violations []Violation `json:"violations"`
}

// Error returns the validation message.
func (err ValidationError) Error() string {
	parts := make([]string, 0, len(err.Violations))
	for _, violation := range err.Violations {
		parts = append(parts, violation.Field+": "+violation.Message)
	}
	return strings.Join(parts, "; ")
}

// AppendViolation appends one validation violation.
func AppendViolation(violations []Violation, field string, message string) []Violation {
	return append(violations, Violation{Field: field, Message: message})
}

// ErrorIfInvalid returns a validation error if violations exist.
func ErrorIfInvalid(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return ValidationError{Violations: violations}
}
