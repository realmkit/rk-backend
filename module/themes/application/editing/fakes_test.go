package editing

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeVersionRepository stores versions for editing tests.
type fakeVersionRepository struct {
	versions map[uuid.UUID]domain.ThemeVersion
}

// Create stores a version.
func (repository *fakeVersionRepository) Create(
	_ context.Context,
	version domain.ThemeVersion,
) (domain.ThemeVersion, error) {
	repository.versions[version.ID] = version
	return version, nil
}

// Update updates a version.
func (repository *fakeVersionRepository) Update(
	_ context.Context,
	version domain.ThemeVersion,
	expectedVersion uint64,
) (domain.ThemeVersion, error) {
	version.Version = expectedVersion + 1
	repository.versions[version.ID] = version
	return version, nil
}

// Archive is unused by editing tests.
func (repository *fakeVersionRepository) Archive(context.Context, uuid.UUID, uint64) error {
	return nil
}

// FindByID returns a version.
func (repository *fakeVersionRepository) FindByID(
	_ context.Context,
	id uuid.UUID,
) (domain.ThemeVersion, error) {
	version, ok := repository.versions[id]
	if !ok {
		return domain.ThemeVersion{}, port.ErrNotFound
	}
	return version, nil
}

// FindBySourceReference is unused by editing tests.
func (repository *fakeVersionRepository) FindBySourceReference(
	context.Context,
	uuid.UUID,
	string,
) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, port.ErrNotFound
}

// ListByTheme is unused by editing tests.
func (repository *fakeVersionRepository) ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeVersion, error) {
	return nil, nil
}

// fakeFileRepository stores files for editing tests.
type fakeFileRepository struct {
	files map[uuid.UUID][]domain.ThemeFile
}

// ReplaceVersionFiles replaces files.
func (repository *fakeFileRepository) ReplaceVersionFiles(
	_ context.Context,
	versionID uuid.UUID,
	files []domain.ThemeFile,
) error {
	repository.files[versionID] = files
	return nil
}

// ListByVersion returns files.
func (repository *fakeFileRepository) ListByVersion(
	_ context.Context,
	versionID uuid.UUID,
) ([]domain.ThemeFile, error) {
	return repository.files[versionID], nil
}

// FindByPath is unused by editing tests.
func (repository *fakeFileRepository) FindByPath(
	context.Context,
	uuid.UUID,
	domain.FilePath,
) (domain.ThemeFile, error) {
	return domain.ThemeFile{}, port.ErrNotFound
}

// fakeAssetRepository is unused by editing tests.
type fakeAssetRepository struct{}

// ReplaceVersionAssets is unused by editing tests.
func (repository fakeAssetRepository) ReplaceVersionAssets(context.Context, uuid.UUID, []domain.ThemeAsset) error {
	return nil
}

// ListByVersion is unused by editing tests.
func (repository fakeAssetRepository) ListByVersion(context.Context, uuid.UUID) ([]domain.ThemeAsset, error) {
	return nil, nil
}

// fakeValidator records revalidation requests.
type fakeValidator struct {
	calls int
}

// Validate records one validation request.
func (validator *fakeValidator) Validate(context.Context, ValidationCommand) (ValidationResult, error) {
	validator.calls++
	return ValidationResult{}, nil
}
