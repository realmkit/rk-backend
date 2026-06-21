package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// benchmarkUploadIntent stores the upload intent benchmark result.
var benchmarkUploadIntent port.UploadIntent

// BenchmarkCreateUploadIntent measures asset normalization, storage key creation, persistence, and presign orchestration.
func BenchmarkCreateUploadIntent(b *testing.B) {
	repository := &benchmarkAssetRepository{}
	service := NewService(repository, &memoryStore{}, "realmkit-assets")
	command := validCommand()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		intent, err := service.CreateUploadIntent(ctx, command)
		if err != nil {
			b.Fatalf("CreateUploadIntent() error = %v", err)
		}
		benchmarkUploadIntent = intent
	}
}

// benchmarkAssetRepository keeps one asset for benchmark service calls.
type benchmarkAssetRepository struct {
	asset domain.Asset
}

// Create stores one benchmark asset.
func (repository *benchmarkAssetRepository) Create(_ context.Context, asset domain.Asset) (domain.Asset, error) {
	repository.asset = asset
	return asset, nil
}

// Update stores one benchmark asset update.
func (repository *benchmarkAssetRepository) Update(
	_ context.Context,
	asset domain.Asset,
	expectedVersion uint64,
) (domain.Asset, error) {
	asset.Version = expectedVersion + 1
	repository.asset = asset
	return asset, nil
}

// FindByID returns the stored benchmark asset.
func (repository *benchmarkAssetRepository) FindByID(context.Context, uuid.UUID) (domain.Asset, error) {
	if repository.asset.ID == uuid.Nil {
		return domain.Asset{}, port.ErrNotFound
	}
	return repository.asset, nil
}

// List returns the stored benchmark asset when present.
func (repository *benchmarkAssetRepository) List(
	context.Context,
	port.AssetFilter,
	pagination.Page,
) (pagination.Result[domain.Asset], error) {
	if repository.asset.ID == uuid.Nil {
		return pagination.Result[domain.Asset]{}, nil
	}
	return pagination.Result[domain.Asset]{Items: []domain.Asset{repository.asset}}, nil
}

// ListNamespaces returns one benchmark namespace.
func (repository *benchmarkAssetRepository) ListNamespaces(context.Context) ([]string, error) {
	return []string{"community"}, nil
}

// ListFolders returns one benchmark folder.
func (repository *benchmarkAssetRepository) ListFolders(context.Context, port.FolderFilter) ([]string, error) {
	return []string{"brand"}, nil
}

// Delete clears the stored benchmark asset.
func (repository *benchmarkAssetRepository) Delete(context.Context, uuid.UUID, uint64) error {
	repository.asset = domain.Asset{}
	return nil
}
