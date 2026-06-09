package port

import "errors"

// ErrNotFound reports that an asset was not found.
var ErrNotFound = errors.New("asset not found")

// ErrConflict reports that an asset conflicts with existing state.
var ErrConflict = errors.New("asset conflict")

// ErrPreconditionFailed reports a stale optimistic version.
var ErrPreconditionFailed = errors.New("asset precondition failed")

// ErrInvalidState reports an invalid asset state transition.
var ErrInvalidState = errors.New("asset invalid state")

// ErrUploadMismatch reports that the stored object does not match intent.
var ErrUploadMismatch = errors.New("asset upload mismatch")
