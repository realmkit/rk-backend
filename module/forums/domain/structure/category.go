package structure

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ForumCategory groups top-level forums for display.
type ForumCategory struct {
	// ID is the category identifier.
	ID uuid.UUID `json:"id"`

	// Key is the stable category key.
	Key Key `json:"key"`

	// Name is the display name.
	Name string `json:"name"`

	// Description explains the category.
	Description string `json:"description"`

	// DisplayOrder controls category ordering.
	DisplayOrder int `json:"display_order"`

	// Status is the category lifecycle state.
	Status CategoryStatus `json:"status"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized category copy.
func (category ForumCategory) Normalize() ForumCategory {
	category.Key = Key(strings.TrimSpace(string(category.Key)))
	category.Name = strings.TrimSpace(category.Name)
	category.Description = strings.TrimSpace(category.Description)
	if category.Status == "" {
		category.Status = CategoryStatusActive
	}
	if category.Version == 0 {
		category.Version = 1
	}
	return category
}

// Validate validates category fields.
func (category ForumCategory) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateKey("key", category.Key)...)
	violations = append(violations, ValidateName("name", category.Name)...)
	violations = append(violations, ValidateDescription("description", category.Description)...)
	violations = append(violations, ValidateDisplayOrder("display_order", category.DisplayOrder)...)
	violations = append(violations, ValidateCategoryStatus("status", category.Status)...)
	return NewValidationError(violations)
}
