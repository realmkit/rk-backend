package domain

import "slices"

// Status is the local RealmKit user lifecycle state.
type Status string

const (
	// StatusActive means the user can authenticate and use RealmKit.
	StatusActive Status = "active"

	// StatusDisabled means the user exists but cannot authenticate.
	StatusDisabled Status = "disabled"

	// StatusPendingProfile means the user exists but local profile setup is incomplete.
	StatusPendingProfile Status = "pending_profile"
)

// ValidateStatus validates user status.
func ValidateStatus(field string, status Status) []Violation {
	if slices.Contains([]Status{StatusActive, StatusDisabled, StatusPendingProfile}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}
