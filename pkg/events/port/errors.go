package port

import "errors"

// ErrNotFound reports that an event does not exist.
var ErrNotFound = errors.New("event not found")

// ErrConflict reports that an event conflicts with existing state.
var ErrConflict = errors.New("event conflict")

// ErrForbidden reports a denied scope subscription.
var ErrForbidden = errors.New("event forbidden")
