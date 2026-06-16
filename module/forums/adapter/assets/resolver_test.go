package assets

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	assetdomain "github.com/realmkit/rk-backend/module/assets/domain"
	assetport "github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestResolverReportsExistingAndMissingAssets verifies asset existence mapping.
func TestResolverReportsExistingAndMissingAssets(t *testing.T) {
	existingID := uuid.New()
	resolver := NewResolver(assetService{assets: map[uuid.UUID]assetdomain.Asset{existingID: {ID: existingID}}})

	exists, err := resolver.AssetExists(context.Background(), existingID)
	if err != nil {
		t.Fatalf("AssetExists(existing) error = %v", err)
	}
	missing, err := resolver.AssetExists(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("AssetExists(missing) error = %v", err)
	}
	if !exists || missing {
		t.Fatalf("exists=%v missing=%v, want true false", exists, missing)
	}
}

// assetService is an asset service test double.
type assetService struct {
	assets map[uuid.UUID]assetdomain.Asset
}

// CreateUploadIntent creates an asset upload intent.
func (service assetService) CreateUploadIntent(context.Context, assetport.CreateUploadIntentCommand) (assetport.UploadIntent, error) {
	return assetport.UploadIntent{}, nil
}

// CompleteUpload completes an asset upload.
func (service assetService) CompleteUpload(context.Context, assetport.CompleteUploadCommand) (assetdomain.Asset, error) {
	return assetdomain.Asset{}, nil
}

// Get returns one asset.
func (service assetService) Get(_ context.Context, id uuid.UUID) (assetdomain.Asset, error) {
	asset, ok := service.assets[id]
	if !ok {
		return assetdomain.Asset{}, assetport.ErrNotFound
	}
	return asset, nil
}

// GetURL returns an asset URL.
func (service assetService) GetURL(context.Context, uuid.UUID, time.Duration) (string, error) {
	return "", nil
}

// List returns assets.
func (service assetService) List(context.Context, assetport.AssetFilter, pagination.Page) (pagination.Result[assetdomain.Asset], error) {
	return pagination.Result[assetdomain.Asset]{}, nil
}

// ListNamespaces returns asset namespaces.
func (service assetService) ListNamespaces(context.Context) ([]string, error) {
	return nil, nil
}

// ListFolders returns asset folders.
func (service assetService) ListFolders(context.Context, assetport.FolderFilter) ([]string, error) {
	return nil, nil
}

// Update updates an asset.
func (service assetService) Update(context.Context, assetport.UpdateAssetCommand) (assetdomain.Asset, error) {
	return assetdomain.Asset{}, nil
}

// Delete deletes an asset.
func (service assetService) Delete(context.Context, assetport.DeleteAssetCommand) error {
	return nil
}
