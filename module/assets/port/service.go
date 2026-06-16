package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Service manages assets.
type Service interface {
	// CreateUploadIntent creates an asset and presigned upload URL.
	CreateUploadIntent(ctx context.Context, command CreateUploadIntentCommand) (UploadIntent, error)

	// CompleteUpload confirms the upload object exists.
	CompleteUpload(ctx context.Context, command CompleteUploadCommand) (domain.Asset, error)

	// Get returns one asset.
	Get(ctx context.Context, id uuid.UUID) (domain.Asset, error)

	// GetURL returns a signed read URL.
	GetURL(ctx context.Context, id uuid.UUID, ttl time.Duration) (string, error)

	// List returns matching assets.
	List(ctx context.Context, filter AssetFilter, page pagination.Page) (pagination.Result[domain.Asset], error)

	// ListNamespaces returns active asset namespaces.
	ListNamespaces(ctx context.Context) ([]string, error)

	// ListFolders returns direct virtual folder children.
	ListFolders(ctx context.Context, filter FolderFilter) ([]string, error)

	// Update changes mutable asset fields.
	Update(ctx context.Context, command UpdateAssetCommand) (domain.Asset, error)

	// Delete soft deletes one asset.
	Delete(ctx context.Context, command DeleteAssetCommand) error
}
