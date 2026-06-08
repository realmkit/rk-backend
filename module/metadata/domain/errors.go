package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalid reports invalid metadata input.
var ErrInvalid = errors.New("invalid metadata")

// Violation describes one validation failure.
type Violation struct {
	// Field is the invalid field path.
	Field string

	// Message explains the validation failure.
	Message string
}

// ValidationError reports one or more validation failures.
type ValidationError struct {
	// Violations contains validation failures.
	Violations []Violation
}

// Error returns a human-readable validation summary.
func (err ValidationError) Error() string {
	if len(err.Violations) == 0 {
		return ErrInvalid.Error()
	}

	parts := make([]string, 0, len(err.Violations))
	for _, violation := range err.Violations {
		parts = append(parts, violation.Field+": "+violation.Message)
	}
	return fmt.Sprintf("%s: %s", ErrInvalid, strings.Join(parts, "; "))
}

// Unwrap returns the base validation error.
func (err ValidationError) Unwrap() error {
	return ErrInvalid
}

// NewValidationError creates a ValidationError from violations.
func NewValidationError(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return ValidationError{Violations: violations}
}

// AppendViolation appends a validation failure.
func AppendViolation(violations []Violation, field string, message string) []Violation {
	return append(violations, Violation{Field: field, Message: message})
}
