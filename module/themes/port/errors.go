package port

import "errors"

var (
	// ErrNotFound reports a missing theme resource.
	ErrNotFound = errors.New("theme resource not found")

	// ErrConflict reports conflicting theme state.
	ErrConflict = errors.New("theme resource conflict")

	// ErrPreconditionFailed reports stale optimistic version state.
	ErrPreconditionFailed = errors.New("theme precondition failed")
)
