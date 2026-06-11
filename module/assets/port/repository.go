package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// AssetRepository stores assets.
type AssetRepository interface {
	// Create stores an asset.
	Create(ctx context.Context, asset domain.Asset) (domain.Asset, error)

	// Update stores mutable asset fields.
	Update(ctx context.Context, asset domain.Asset, expectedVersion uint64) (domain.Asset, error)

	// FindByID returns one asset.
	FindByID(ctx context.Context, id uuid.UUID) (domain.Asset, error)

	// List returns matching assets.
	List(ctx context.Context, filter AssetFilter, page pagination.Page) (pagination.Result[domain.Asset], error)

	// ListFolders returns direct child folders.
	ListFolders(ctx context.Context, filter FolderFilter) ([]string, error)

	// Delete soft deletes one asset.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error
}
