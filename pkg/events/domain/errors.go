package domain

import "strings"

// Violation describes one event validation failure.
type Violation struct {
	// Field is the invalid field name.
	Field string `json:"field"`

	// Message describes the validation failure.
	Message string `json:"message"`
}

// ValidationError reports invalid event data.
type ValidationError struct {
	// Violations contains all detected validation failures.
	Violations []Violation `json:"violations"`
}

// Error returns the validation error message.
func (err ValidationError) Error() string {
	messages := make([]string, 0, len(err.Violations))
	for _, violation := range err.Violations {
		messages = append(messages, violation.Field+": "+violation.Message)
	}
	return strings.Join(messages, "; ")
}

// AppendViolation appends one validation violation.
func AppendViolation(violations []Violation, field string, message string) []Violation {
	return append(violations, Violation{Field: field, Message: message})
}

// ErrorIfInvalid returns a validation error when violations exist.
func ErrorIfInvalid(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return ValidationError{Violations: violations}
}
