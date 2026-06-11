package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Group organizes users for display and authorization relations.
type Group struct {
	// ID is the group identifier.
	ID uuid.UUID `json:"id"`

	// Key is the stable lower snake key.
	Key Key `json:"key"`

	// Name is the human-facing name.
	Name string `json:"name"`

	// Description explains the group purpose.
	Description string `json:"description"`

	// Color is the frontend display color.
	Color Color `json:"color"`

	// Weight helps the frontend choose a display group.
	Weight int `json:"weight"`

	// Status is the group lifecycle state.
	Status GroupStatus `json:"status"`

	// IconAssetID is the optional icon asset identifier.
	IconAssetID *uuid.UUID `json:"icon_asset_id,omitempty"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates group fields.
func (group Group) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateKey("key", group.Key)...)
	violations = append(violations, validateName("name", group.Name)...)
	if len(group.Description) > 500 {
		violations = AppendViolation(violations, "description", "must be at most 500 characters")
	}
	violations = append(violations, ValidateColor("color", group.Color)...)
	if group.Weight < 0 {
		violations = AppendViolation(violations, "weight", "must be zero or greater")
	}
	violations = append(violations, ValidateGroupStatus("status", group.Status)...)
	return NewValidationError(violations)
}

// GrantsPermissions reports whether group status can grant permissions.
func (group Group) GrantsPermissions() bool {
	return group.Status == GroupStatusActive || group.Status == GroupStatusSystem
}

// validateName validates human-facing names.
func validateName(field string, value string) []Violation {
	value = strings.TrimSpace(value)
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if len(value) > 120 {
		return []Violation{{Field: field, Message: "must be at most 120 characters"}}
	}
	return nil
}

// DisplayGroup returns the frontend display group from active memberships.
func DisplayGroup(groups []Group, memberships []Membership, instant time.Time) (Group, bool) {
	byID := map[uuid.UUID]Group{}
	for _, group := range groups {
		if group.GrantsPermissions() {
			byID[group.ID] = group
		}
	}
	var selected Group
	var selectedMembership Membership
	found := false
	for _, membership := range memberships {
		group, ok := byID[membership.GroupID]
		if !ok || !membership.ActiveAt(instant) {
			continue
		}
		if !found || betterDisplayGroup(group, membership, selected, selectedMembership) {
			selected = group
			selectedMembership = membership
			found = true
		}
	}
	return selected, found
}

// betterDisplayGroup reports whether candidate should be displayed first.
func betterDisplayGroup(
	candidate Group,
	candidateMembership Membership,
	current Group,
	currentMembership Membership,
) bool {
	if candidate.Weight != current.Weight {
		return candidate.Weight > current.Weight
	}
	if !candidateMembership.CreatedAt.Equal(currentMembership.CreatedAt) {
		return candidateMembership.CreatedAt.Before(currentMembership.CreatedAt)
	}
	return candidate.Key < current.Key
}
