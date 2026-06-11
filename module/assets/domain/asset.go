package domain

import (
	"time"

	"github.com/google/uuid"
)

// Asset stores one logical file managed by RealmKit.
type Asset struct {
	// ID is the asset identifier.
	ID uuid.UUID `json:"id"`

	// Namespace is the browsable namespace.
	Namespace Namespace `json:"namespace"`

	// Path is the virtual folder path.
	Path VirtualPath `json:"path"`

	// Filename is the stored filename.
	Filename Filename `json:"filename"`

	// DisplayName is the human-readable name.
	DisplayName string `json:"display_name"`

	// Visibility describes URL resolution visibility.
	Visibility Visibility `json:"visibility"`

	// Status describes upload and moderation state.
	Status Status `json:"status"`

	// StorageKey is the object key in S3.
	StorageKey string `json:"storage_key"`

	// Bucket is the storage bucket.
	Bucket string `json:"bucket"`

	// ContentType is the media type.
	ContentType string `json:"content_type"`

	// SizeBytes is the expected or confirmed object size.
	SizeBytes int64 `json:"size_bytes"`

	// ETag is the storage entity tag.
	ETag string `json:"etag,omitempty"`

	// CreatedByUserID is the user that created the asset when known.
	CreatedByUserID *uuid.UUID `json:"created_by_user_id,omitempty"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates asset fields.
func (asset Asset) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateNamespace("namespace", asset.Namespace)...)
	violations = append(violations, ValidatePath("path", asset.Path)...)
	violations = append(violations, ValidateFilename("filename", asset.Filename)...)
	violations = append(violations, ValidateVisibility("visibility", asset.Visibility)...)
	violations = append(violations, ValidateStatus("status", asset.Status)...)
	violations = append(violations, ValidateContentType("content_type", asset.ContentType)...)
	if asset.SizeBytes <= 0 {
		violations = AppendViolation(violations, "size_bytes", "must be greater than zero")
	}
	if asset.StorageKey == "" {
		violations = AppendViolation(violations, "storage_key", "is required")
	}
	if asset.Bucket == "" {
		violations = AppendViolation(violations, "bucket", "is required")
	}
	return NewValidationError(violations)
}

// Normalize returns a normalized asset copy.
func (asset Asset) Normalize() Asset {
	asset.Path = NormalizePath(asset.Path)
	asset.Filename = NormalizeFilename(asset.Filename)
	if asset.DisplayName == "" {
		asset.DisplayName = string(asset.Filename)
	}
	return asset
}
