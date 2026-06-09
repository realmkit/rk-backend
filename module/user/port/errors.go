package port

import "errors"

// ErrNotFound reports that a user resource was not found.
var ErrNotFound = errors.New("user resource not found")

// ErrConflict reports a user resource conflict.
var ErrConflict = errors.New("user resource conflict")

// ErrPreconditionFailed reports a stale optimistic version.
var ErrPreconditionFailed = errors.New("user precondition failed")

// ErrDisabled reports that the local user cannot authenticate.
var ErrDisabled = errors.New("user disabled")
