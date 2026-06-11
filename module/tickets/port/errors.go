// Package port defines ticket application contracts.
package port

import "errors"

var (
	// ErrNotFound reports a missing ticket resource.
	ErrNotFound = errors.New("ticket not found")
	// ErrConflict reports conflicting ticket state.
	ErrConflict = errors.New("ticket conflict")
	// ErrForbidden reports denied ticket access.
	ErrForbidden = errors.New("ticket forbidden")
	// ErrPreconditionFailed reports stale optimistic version.
	ErrPreconditionFailed = errors.New("ticket precondition failed")
)
