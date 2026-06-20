package validation

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeVersionRepository stores one validation version.
type fakeVersionRepository struct {
	version domain.ThemeVersion
}

// Create is unused by validation tests.
func (repository *fakeVersionRepository) Create(
	context.Context,
	domain.ThemeVersion,
) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, nil
}

// Update records the updated version.
func (repository *fakeVersionRepository) Update(
	_ context.Context,
	version domain.ThemeVersion,
	expectedVersion uint64,
) (domain.ThemeVersion, error) {
	version.Version = expectedVersion + 1
	repository.version = version
	return version, nil
}

// Archive is unused by validation tests.
func (repository *fakeVersionRepository) Archive(context.Context, uuid.UUID, uint64) error {
	return nil
}

// FindByID returns the stored version.
func (repository *fakeVersionRepository) FindByID(
	context.Context,
	uuid.UUID,
) (domain.ThemeVersion, error) {
	return repository.version, nil
}

// FindBySourceReference is unused by validation tests.
func (repository *fakeVersionRepository) FindBySourceReference(
	context.Context,
	uuid.UUID,
	string,
) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, port.ErrNotFound
}

// ListByTheme is unused by validation tests.
func (repository *fakeVersionRepository) ListByTheme(
	context.Context,
	uuid.UUID,
) ([]domain.ThemeVersion, error) {
	return nil, nil
}

// fakeFileRepository stores validation files.
type fakeFileRepository struct {
	files []domain.ThemeFile
}

// ReplaceVersionFiles records files.
func (repository *fakeFileRepository) ReplaceVersionFiles(
	context.Context,
	uuid.UUID,
	[]domain.ThemeFile,
) error {
	return nil
}

// ListByVersion returns files.
func (repository *fakeFileRepository) ListByVersion(
	context.Context,
	uuid.UUID,
) ([]domain.ThemeFile, error) {
	return repository.files, nil
}

// FindByPath is unused by validation tests.
func (repository *fakeFileRepository) FindByPath(
	context.Context,
	uuid.UUID,
	domain.FilePath,
) (domain.ThemeFile, error) {
	return domain.ThemeFile{}, port.ErrNotFound
}

// fakeIssueRepository records validation issues.
type fakeIssueRepository struct {
	issues []domain.ThemeValidationIssue
}

// ReplaceVersionIssues records issues.
func (repository *fakeIssueRepository) ReplaceVersionIssues(
	_ context.Context,
	_ uuid.UUID,
	issues []domain.ThemeValidationIssue,
) error {
	repository.issues = issues
	return nil
}

// ListByVersion returns issues.
func (repository *fakeIssueRepository) ListByVersion(
	context.Context,
	uuid.UUID,
) ([]domain.ThemeValidationIssue, error) {
	return repository.issues, nil
}
