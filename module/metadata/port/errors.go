package port

import "errors"

// ErrNotFound reports that a metadata resource does not exist.
var ErrNotFound = errors.New("metadata resource not found")

// ErrConflict reports that a metadata resource conflicts with existing state.
var ErrConflict = errors.New("metadata conflict")

// ErrPreconditionFailed reports that an optimistic concurrency precondition failed.
var ErrPreconditionFailed = errors.New("metadata precondition failed")

// ErrInactive reports that a metadata resource is inactive.
var ErrInactive = errors.New("metadata resource inactive")

// ErrReferenced reports that a metadata resource still has active dependents.
var ErrReferenced = errors.New("metadata resource referenced")
