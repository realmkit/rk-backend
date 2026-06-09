package domain

import "errors"

// ErrInvalid reports invalid forum domain data.
var ErrInvalid = errors.New("invalid forum data")

// Violation describes one validation failure.
type Violation struct {
	// Field is the invalid field path.
	Field string `json:"field"`

	// Message describes the validation failure.
	Message string `json:"message"`
}

// ValidationError contains domain validation failures.
type ValidationError struct {
	// Violations contains every validation failure.
	Violations []Violation `json:"errors"`
}

// Error returns the validation error message.
func (err ValidationError) Error() string {
	return ErrInvalid.Error()
}

// Unwrap returns the sentinel invalid error.
func (err ValidationError) Unwrap() error {
	return ErrInvalid
}

// NewValidationError returns nil when violations is empty.
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
