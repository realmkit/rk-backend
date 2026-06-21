package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/gorm"
)

// ThemeRepository stores theme families in PostgreSQL.
type ThemeRepository struct {
	store orm.Store // store stores the store value.
}

// NewThemeRepository creates a theme repository.
func NewThemeRepository(store orm.Store) ThemeRepository {
	return ThemeRepository{store: store}
}

// Create stores a theme family.
func (repository ThemeRepository) Create(ctx context.Context, theme domain.Theme) (domain.Theme, error) {
	model := themeModel(theme)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Theme{}, port.ErrConflict
	}
	return themeFromModel(model), nil
}

// Update stores mutable theme fields.
func (repository ThemeRepository) Update(
	ctx context.Context,
	theme domain.Theme,
	expectedVersion uint64,
) (domain.Theme, error) {
	result := repository.store.DB(ctx).Model(&ThemeModel{}).
		Where("id = ? AND version = ?", theme.ID, expectedVersion).
		Updates(map[string]any{
			"key":                string(theme.Key),
			"name":               theme.Name,
			"description":        theme.Description,
			"status":             string(theme.Status),
			"updated_by_user_id": theme.UpdatedBy,
			"version":            expectedVersion + 1,
		})
	if result.Error != nil {
		return domain.Theme{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Theme{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, theme.ID)
}

// Archive marks a theme family as archived.
func (repository ThemeRepository) Archive(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Model(&ThemeModel{}).
		Where("id = ? AND version = ?", id, expectedVersion).
		Updates(map[string]any{"status": string(domain.ThemeStatusArchived), "version": expectedVersion + 1})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// FindByID returns one theme family.
func (repository ThemeRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Theme, error) {
	var model ThemeModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Theme{}, mapError(err)
	}
	return themeFromModel(model), nil
}

// List returns matching theme families.
func (repository ThemeRepository) List(ctx context.Context, filter port.ThemeFilter) ([]domain.Theme, error) {
	query := repository.store.DB(ctx).Model(&ThemeModel{}).Order("name ASC, id ASC")
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var models []ThemeModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	themes := make([]domain.Theme, 0, len(models))
	for _, model := range models {
		themes = append(themes, themeFromModel(model))
	}
	return themes, nil
}

// VersionRepository stores theme versions in PostgreSQL.
type VersionRepository struct {
	store orm.Store // store stores the store value.
}

// NewVersionRepository creates a theme version repository.
func NewVersionRepository(store orm.Store) VersionRepository {
	return VersionRepository{store: store}
}

// Create stores one theme version.
func (repository VersionRepository) Create(
	ctx context.Context,
	version domain.ThemeVersion,
) (domain.ThemeVersion, error) {
	model := versionModel(version)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.ThemeVersion{}, port.ErrConflict
	}
	return versionFromModel(model), nil
}

// Update stores mutable version metadata.
func (repository VersionRepository) Update(
	ctx context.Context,
	version domain.ThemeVersion,
	expectedVersion uint64,
) (domain.ThemeVersion, error) {
	current, err := repository.FindByID(ctx, version.ID)
	if err != nil {
		return domain.ThemeVersion{}, err
	}
	if err := current.EnsureEditable(); err != nil {
		return domain.ThemeVersion{}, err
	}
	result := repository.store.DB(ctx).Model(&VersionModel{}).
		Where("id = ? AND version = ?", version.ID, expectedVersion).
		Updates(versionUpdates(version, expectedVersion))
	if result.Error != nil {
		return domain.ThemeVersion{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ThemeVersion{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, version.ID)
}

// Archive marks a version as archived.
func (repository VersionRepository) Archive(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Model(&VersionModel{}).
		Where("id = ? AND version = ?", id, expectedVersion).
		Updates(map[string]any{"status": string(domain.VersionStatusArchived), "version": expectedVersion + 1})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// FindByID returns one version.
func (repository VersionRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.ThemeVersion, error) {
	var model VersionModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.ThemeVersion{}, mapError(err)
	}
	return versionFromModel(model), nil
}

// FindBySourceReference returns one version by retry-safe source reference.
func (repository VersionRepository) FindBySourceReference(
	ctx context.Context,
	themeID uuid.UUID,
	sourceReference string,
) (domain.ThemeVersion, error) {
	var model VersionModel
	err := repository.store.DB(ctx).
		First(&model, "theme_id = ? AND source_reference = ?", themeID, sourceReference).
		Error
	if err != nil {
		return domain.ThemeVersion{}, mapError(err)
	}
	return versionFromModel(model), nil
}

// ListByTheme returns versions for a theme family.
func (repository VersionRepository) ListByTheme(ctx context.Context, themeID uuid.UUID) ([]domain.ThemeVersion, error) {
	var models []VersionModel
	err := repository.store.DB(ctx).
		Where("theme_id = ?", themeID).
		Order("created_at DESC, id ASC").
		Find(&models).
		Error
	if err != nil {
		return nil, err
	}
	versions := make([]domain.ThemeVersion, 0, len(models))
	for _, model := range models {
		versions = append(versions, versionFromModel(model))
	}
	return versions, nil
}

// versionUpdates returns persistence updates for one version.
func versionUpdates(version domain.ThemeVersion, expectedVersion uint64) map[string]any {
	return map[string]any{
		"semver":               version.Semver,
		"label":                version.Label,
		"status":               string(version.Status),
		"source_kind":          string(version.SourceKind),
		"source_reference":     version.SourceReference,
		"package_storage_key":  version.PackageStorageKey,
		"package_size_bytes":   version.PackageSizeBytes,
		"manifest_json":        jsonString(version.ManifestJSON),
		"settings_schema_json": jsonString(version.SettingsSchemaJSON),
		"settings_data_json":   jsonString(version.SettingsDataJSON),
		"integrity_sha256":     string(version.IntegritySHA256),
		"published_at":         version.PublishedAt,
		"published_by_user_id": version.PublishedBy,
		"updated_by_user_id":   version.UpdatedBy,
		"version":              expectedVersion + 1,
	}
}

// mapError maps persistence errors to theme port errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, orm.ErrNotFound) {
		return port.ErrNotFound
	}
	if errors.Is(err, orm.ErrConflict) {
		return port.ErrConflict
	}
	return err
}

// Ensure repositories implement their ports.
var (
	_ port.ThemeRepository   = ThemeRepository{}
	_ port.VersionRepository = VersionRepository{}
)
