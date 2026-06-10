package port

import "errors"

// ErrNotFound reports a missing cron object.
var ErrNotFound = errors.New("cronjob not found")

// ErrPreconditionFailed reports a stale version.
var ErrPreconditionFailed = errors.New("cronjob precondition failed")

// ErrNoDueJob reports that no job is due.
var ErrNoDueJob = errors.New("no cronjob due")

// ErrHandlerMissing reports that no handler was registered.
var ErrHandlerMissing = errors.New("cronjob handler missing")
