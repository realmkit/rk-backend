package domain

import "errors"

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
