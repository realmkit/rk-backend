package port

import "errors"

// ErrNotFound reports that a forum resource was not found.
var ErrNotFound = errors.New("forum resource not found")

// ErrConflict reports a conflicting forum state.
var ErrConflict = errors.New("forum resource conflict")

// ErrPreconditionFailed reports a stale optimistic version.
var ErrPreconditionFailed = errors.New("forum precondition failed")

// ErrForbidden reports a denied forum permission.
var ErrForbidden = errors.New("forum permission denied")

// ErrInvalidMove reports an invalid forum tree move.
var ErrInvalidMove = errors.New("invalid forum move")
