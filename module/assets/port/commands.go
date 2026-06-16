package port

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/pkg/search"
)

// CreateUploadIntentCommand requests an asset row and signed upload URL.
type CreateUploadIntentCommand struct {
	// Namespace is the browsable namespace.
	Namespace domain.Namespace

	// Path is the virtual folder path.
	Path domain.VirtualPath

	// Filename is the plain filename.
	Filename domain.Filename

	// DisplayName is the optional human-readable name.
	DisplayName string

	// Visibility controls future URL resolution.
	Visibility domain.Visibility

	// ContentType is the upload media type.
	ContentType string

	// SizeBytes is the expected object size.
	SizeBytes int64

	// CreatedByUserID is the user creating the intent when known.
	CreatedByUserID *uuid.UUID
}

// UpdateAssetCommand changes mutable asset metadata.
type UpdateAssetCommand struct {
	// ID is the asset identifier.
	ID uuid.UUID

	// Namespace is the replacement namespace.
	Namespace domain.Namespace

	// DisplayName is the replacement display name.
	DisplayName string

	// Path is the replacement virtual folder path.
	Path domain.VirtualPath

	// Visibility is the replacement visibility.
	Visibility domain.Visibility

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// CompleteUploadCommand confirms the upload exists in storage.
type CompleteUploadCommand struct {
	// ID is the asset identifier.
	ID uuid.UUID
}

// DeleteAssetCommand soft deletes an asset.
type DeleteAssetCommand struct {
	// ID is the asset identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// AssetFilter filters asset lists.
type AssetFilter struct {
	// Namespace filters by namespace.
	Namespace domain.Namespace

	// Path filters by exact virtual folder path.
	Path domain.VirtualPath

	// PathExact reports whether Path should filter even when it is empty.
	PathExact bool

	// PathPrefix filters by folder prefix.
	PathPrefix domain.VirtualPath

	// Status filters by status.
	Status domain.Status

	// Visibility filters by asset visibility.
	Visibility domain.Visibility

	// Query filters by filename, display name, or path.
	Query search.TextQuery

	// Sort controls deterministic result ordering.
	Sort search.Sort
}

// DefaultAssetSort returns the default asset list sort.
func DefaultAssetSort() search.SortOption {
	return search.SortOption{Key: "created_at", DefaultDirection: search.DirectionDesc}
}

// AllowedAssetSorts returns public asset list sort keys.
func AllowedAssetSorts() []search.SortOption {
	return []search.SortOption{
		DefaultAssetSort(),
		{Key: "filename", DefaultDirection: search.DirectionAsc},
		{Key: "display_name", DefaultDirection: search.DirectionAsc},
		{Key: "updated_at", DefaultDirection: search.DirectionDesc},
	}
}

// FolderFilter filters virtual folder lists.
type FolderFilter struct {
	// Namespace filters by namespace.
	Namespace domain.Namespace

	// PathPrefix filters by parent virtual folder.
	PathPrefix domain.VirtualPath
}

// UploadIntent contains an asset and a signed upload request.
type UploadIntent struct {
	// Asset is the pending asset.
	Asset domain.Asset

	// Method is the upload HTTP method.
	Method string

	// URL is the presigned upload URL.
	URL string

	// Headers contains required upload headers.
	Headers map[string]string

	// ExpiresAt is when the URL expires.
	ExpiresAt time.Time
}
