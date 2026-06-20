package delivery

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeThemeRepository stores themes in memory.
type fakeThemeRepository struct {
	themes map[uuid.UUID]domain.Theme
}

// Create stores one theme.
func (repository fakeThemeRepository) Create(context.Context, domain.Theme) (domain.Theme, error) {
	return domain.Theme{}, nil
}

// Update updates one theme.
func (repository fakeThemeRepository) Update(context.Context, domain.Theme, uint64) (domain.Theme, error) {
	return domain.Theme{}, nil
}

// Archive archives one theme.
func (repository fakeThemeRepository) Archive(context.Context, uuid.UUID, uint64) error {
	return nil
}

// FindByID returns one theme.
func (repository fakeThemeRepository) FindByID(_ context.Context, id uuid.UUID) (domain.Theme, error) {
	theme, ok := repository.themes[id]
	if !ok {
		return domain.Theme{}, port.ErrNotFound
	}
	return theme, nil
}

// List returns themes.
func (repository fakeThemeRepository) List(context.Context, port.ThemeFilter) ([]domain.Theme, error) {
	return nil, nil
}

// fakeVersionRepository stores versions in memory.
type fakeVersionRepository struct {
	versions map[uuid.UUID]domain.ThemeVersion
}

// Create stores one version.
func (repository fakeVersionRepository) Create(context.Context, domain.ThemeVersion) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, nil
}

// Update updates one version.
func (repository fakeVersionRepository) Update(context.Context, domain.ThemeVersion, uint64) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, nil
}

// Archive archives one version.
func (repository fakeVersionRepository) Archive(context.Context, uuid.UUID, uint64) error {
	return nil
}

// FindByID returns one version.
func (repository fakeVersionRepository) FindByID(_ context.Context, id uuid.UUID) (domain.ThemeVersion, error) {
	version, ok := repository.versions[id]
	if !ok {
		return domain.ThemeVersion{}, port.ErrNotFound
	}
	return version, nil
}

// FindBySourceReference returns no version.
func (repository fakeVersionRepository) FindBySourceReference(context.Context, uuid.UUID, string) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, port.ErrNotFound
}

// ListByTheme returns no versions.
func (repository fakeVersionRepository) ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeVersion, error) {
	return nil, nil
}

// fakeFileRepository stores files in memory.
type fakeFileRepository struct {
	files map[uuid.UUID][]domain.ThemeFile
}

// ReplaceVersionFiles replaces files.
func (repository fakeFileRepository) ReplaceVersionFiles(context.Context, uuid.UUID, []domain.ThemeFile) error {
	return nil
}

// ListByVersion returns files.
func (repository fakeFileRepository) ListByVersion(_ context.Context, versionID uuid.UUID) ([]domain.ThemeFile, error) {
	return repository.files[versionID], nil
}

// FindByPath returns one file.
func (repository fakeFileRepository) FindByPath(
	_ context.Context,
	versionID uuid.UUID,
	filePath domain.FilePath,
) (domain.ThemeFile, error) {
	for _, file := range repository.files[versionID] {
		if domain.NormalizeFilePath(file.Path) == domain.NormalizeFilePath(filePath) {
			return file, nil
		}
	}
	return domain.ThemeFile{}, port.ErrNotFound
}

// fakeAssetRepository stores assets in memory.
type fakeAssetRepository struct {
	assets map[uuid.UUID][]domain.ThemeAsset
}

// ReplaceVersionAssets replaces assets.
func (repository fakeAssetRepository) ReplaceVersionAssets(context.Context, uuid.UUID, []domain.ThemeAsset) error {
	return nil
}

// ListByVersion returns assets.
func (repository fakeAssetRepository) ListByVersion(_ context.Context, versionID uuid.UUID) ([]domain.ThemeAsset, error) {
	return repository.assets[versionID], nil
}
