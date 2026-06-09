package port

import "errors"

// ErrNotFound reports that a group resource was not found.
var ErrNotFound = errors.New("group resource not found")

// ErrConflict reports a conflicting group state.
var ErrConflict = errors.New("group resource conflict")

// ErrPreconditionFailed reports a stale optimistic version.
var ErrPreconditionFailed = errors.New("group precondition failed")

// ErrForbidden reports a denied permission decision.
var ErrForbidden = errors.New("group permission denied")

// ErrUnknownPermission reports an unknown permission.
var ErrUnknownPermission = errors.New("unknown permission")
