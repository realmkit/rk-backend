package publication

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeVersionRepository stores publication versions.
type fakeVersionRepository struct {
	versions map[uuid.UUID]domain.ThemeVersion
}

// Create stores a version.
func (repository *fakeVersionRepository) Create(
	context.Context,
	domain.ThemeVersion,
) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, nil
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

// Archive is unused by publication tests.
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

// FindBySourceReference is unused by publication tests.
func (repository *fakeVersionRepository) FindBySourceReference(
	context.Context,
	uuid.UUID,
	string,
) (domain.ThemeVersion, error) {
	return domain.ThemeVersion{}, port.ErrNotFound
}

// ListByTheme is unused by publication tests.
func (repository *fakeVersionRepository) ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeVersion, error) {
	return nil, nil
}

// fakeIssueRepository stores validation issues.
type fakeIssueRepository struct {
	issues map[uuid.UUID][]domain.ThemeValidationIssue
}

// ReplaceVersionIssues is unused by publication tests.
func (repository *fakeIssueRepository) ReplaceVersionIssues(
	context.Context,
	uuid.UUID,
	[]domain.ThemeValidationIssue,
) error {
	return nil
}

// ListByVersion returns validation issues.
func (repository *fakeIssueRepository) ListByVersion(
	_ context.Context,
	versionID uuid.UUID,
) ([]domain.ThemeValidationIssue, error) {
	return repository.issues[versionID], nil
}

// fakeSignatureRepository stores package signatures.
type fakeSignatureRepository struct {
	signatures map[uuid.UUID]domain.ThemePackageSignature
}

// ReplaceVersionSignature is unused by publication tests.
func (repository *fakeSignatureRepository) ReplaceVersionSignature(
	context.Context,
	uuid.UUID,
	domain.ThemePackageSignature,
) error {
	return nil
}

// FindByVersion returns a package signature.
func (repository *fakeSignatureRepository) FindByVersion(
	_ context.Context,
	versionID uuid.UUID,
) (domain.ThemePackageSignature, error) {
	signature, ok := repository.signatures[versionID]
	if !ok {
		return domain.ThemePackageSignature{}, port.ErrNotFound
	}
	return signature, nil
}
