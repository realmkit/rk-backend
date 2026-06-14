package postgres

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// AssetRepository stores assets in PostgreSQL.
type AssetRepository struct {
	store orm.Store
}

// NewAssetRepository creates an asset repository.
func NewAssetRepository(store orm.Store) AssetRepository {
	return AssetRepository{store: store}
}

// Create stores an asset.
func (repository AssetRepository) Create(ctx context.Context, asset domain.Asset) (domain.Asset, error) {
	model := assetModelFromDomain(asset)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Asset{}, port.ErrConflict
	}
	return assetFromModel(model), nil
}

// Update stores mutable asset fields.
func (repository AssetRepository) Update(ctx context.Context, asset domain.Asset, expectedVersion uint64) (domain.Asset, error) {
	result := repository.store.DB(ctx).
		Model(&AssetModel{}).
		Where("id = ? AND version = ?", asset.ID, expectedVersion).
		Updates(map[string]any{
			"path":         string(asset.Path),
			"display_name": asset.DisplayName,
			"visibility":   string(asset.Visibility),
			"status":       string(asset.Status),
			"etag":         asset.ETag,
			"version":      expectedVersion + 1,
		})
	if result.Error != nil {
		return domain.Asset{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Asset{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, asset.ID)
}

// FindByID returns one asset.
func (repository AssetRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Asset, error) {
	var model AssetModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Asset{}, mapError(err)
	}
	return assetFromModel(model), nil
}

// List returns matching assets.
func (repository AssetRepository) List(
	ctx context.Context,
	filter port.AssetFilter,
	page pagination.Page,
) (pagination.Result[domain.Asset], error) {
	sort := filter.Sort
	if sort.Key == "" {
		sort, _ = search.NewSort("", "", port.DefaultAssetSort(), port.AllowedAssetSorts())
	}
	filterHash := assetFilterHash(filter, sort)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, sort)
	if err != nil {
		return pagination.Result[domain.Asset]{}, err
	}
	query := applyAssetFilter(repository.store.DB(ctx).Model(&AssetModel{}), filter)
	query, err = applyAssetCursor(query, cursor, hasCursor, sort)
	if err != nil {
		return pagination.Result[domain.Asset]{}, err
	}
	query = query.Order(assetOrder(sort)).Limit(page.Limit + 1)
	var models []AssetModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Asset]{}, err
	}
	return assetPage(models, page.Limit, filterHash, sort)
}

// ListFolders returns direct child folders.
func (repository AssetRepository) ListFolders(ctx context.Context, filter port.FolderFilter) ([]string, error) {
	query := repository.store.DB(ctx).Model(&AssetModel{}).Select("path").Where("namespace = ?", filter.Namespace)
	if filter.PathPrefix != "" {
		query = query.Where("path = ? OR path LIKE ?", filter.PathPrefix, string(filter.PathPrefix)+"/%")
	}
	var paths []string
	if err := query.Pluck("path", &paths).Error; err != nil {
		return nil, err
	}
	return foldersFromPaths(paths, string(filter.PathPrefix)), nil
}

// Delete soft deletes one asset.
func (repository AssetRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&AssetModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// foldersFromPaths returns direct child folder names.
func foldersFromPaths(paths []string, prefix string) []string {
	seen := map[string]struct{}{}
	base := strings.Trim(prefix, "/")
	for _, current := range paths {
		remainder := strings.Trim(current, "/")
		if base != "" {
			remainder = strings.TrimPrefix(remainder, base)
			remainder = strings.Trim(remainder, "/")
		}
		if remainder == "" {
			continue
		}
		seen[strings.Split(remainder, "/")[0]] = struct{}{}
	}
	folders := make([]string, 0, len(seen))
	for folder := range seen {
		folders = append(folders, folder)
	}
	sort.Strings(folders)
	return folders
}

// mapError maps GORM errors into assets errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}

// Ensure AssetRepository implements port.AssetRepository.
var _ port.AssetRepository = AssetRepository{}
