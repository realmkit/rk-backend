package importing

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/application/signing"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeRepositories groups importer repository fakes.
type fakeRepositories struct {
	versions   *fakeVersionRepository
	files      *fakeFileRepository
	assets     *fakeAssetRepository
	issues     *fakeIssueRepository
	signatures *fakeSignatureRepository
	keys       fakeSigningKeyRepository
}

// newFakeRepositories creates importer repository fakes.
func newFakeRepositories() fakeRepositories {
	return fakeRepositories{
		versions:   &fakeVersionRepository{versions: map[uuid.UUID]domain.ThemeVersion{}},
		files:      &fakeFileRepository{files: map[uuid.UUID][]domain.ThemeFile{}},
		assets:     &fakeAssetRepository{assets: map[uuid.UUID][]domain.ThemeAsset{}},
		issues:     &fakeIssueRepository{issues: map[uuid.UUID][]domain.ThemeValidationIssue{}},
		signatures: &fakeSignatureRepository{signatures: map[uuid.UUID]domain.ThemePackageSignature{}},
		keys:       fakeSigningKeyRepository{},
	}
}

// ports returns importer repository ports.
func (repositories fakeRepositories) ports() Repositories {
	return Repositories{
		Versions:    repositories.versions,
		Files:       repositories.files,
		Assets:      repositories.assets,
		Issues:      repositories.issues,
		Signatures:  repositories.signatures,
		SigningKeys: repositories.keys,
	}
}

// fakeVersionRepository stores versions in memory.
type fakeVersionRepository struct {
	versions map[uuid.UUID]domain.ThemeVersion
}

// Create stores one version.
func (repository *fakeVersionRepository) Create(
	_ context.Context,
	version domain.ThemeVersion,
) (domain.ThemeVersion, error) {
	if version.Version == 0 {
		version.Version = 1
	}
	repository.versions[version.ID] = version
	return version, nil
}

// Update updates one version.
func (repository *fakeVersionRepository) Update(
	_ context.Context,
	version domain.ThemeVersion,
	expectedVersion uint64,
) (domain.ThemeVersion, error) {
	version.Version = expectedVersion + 1
	repository.versions[version.ID] = version
	return version, nil
}

// Archive is unused by importer tests.
func (repository *fakeVersionRepository) Archive(context.Context, uuid.UUID, uint64) error {
	return nil
}

// FindByID returns one version.
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

// FindBySourceReference returns one idempotent version.
func (repository *fakeVersionRepository) FindBySourceReference(
	_ context.Context,
	themeID uuid.UUID,
	sourceReference string,
) (domain.ThemeVersion, error) {
	for _, version := range repository.versions {
		if version.ThemeID == themeID && version.SourceReference == sourceReference {
			return version, nil
		}
	}
	return domain.ThemeVersion{}, port.ErrNotFound
}

// ListByTheme is unused by importer tests.
func (repository *fakeVersionRepository) ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeVersion, error) {
	return nil, nil
}

// fakeFileRepository stores files in memory.
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

// FindByPath is unused by importer tests.
func (repository *fakeFileRepository) FindByPath(
	context.Context,
	uuid.UUID,
	domain.FilePath,
) (domain.ThemeFile, error) {
	return domain.ThemeFile{}, port.ErrNotFound
}

// fakeAssetRepository stores assets in memory.
type fakeAssetRepository struct {
	assets map[uuid.UUID][]domain.ThemeAsset
}

// ReplaceVersionAssets replaces assets.
func (repository *fakeAssetRepository) ReplaceVersionAssets(
	_ context.Context,
	versionID uuid.UUID,
	assets []domain.ThemeAsset,
) error {
	repository.assets[versionID] = assets
	return nil
}

// ListByVersion returns assets.
func (repository *fakeAssetRepository) ListByVersion(
	_ context.Context,
	versionID uuid.UUID,
) ([]domain.ThemeAsset, error) {
	return repository.assets[versionID], nil
}

// fakeIssueRepository stores issues in memory.
type fakeIssueRepository struct {
	issues map[uuid.UUID][]domain.ThemeValidationIssue
}

// ReplaceVersionIssues replaces issues.
func (repository *fakeIssueRepository) ReplaceVersionIssues(
	_ context.Context,
	versionID uuid.UUID,
	issues []domain.ThemeValidationIssue,
) error {
	repository.issues[versionID] = issues
	return nil
}

// ListByVersion returns issues.
func (repository *fakeIssueRepository) ListByVersion(
	_ context.Context,
	versionID uuid.UUID,
) ([]domain.ThemeValidationIssue, error) {
	return repository.issues[versionID], nil
}

// fakeSignatureRepository stores signatures in memory.
type fakeSignatureRepository struct {
	signatures map[uuid.UUID]domain.ThemePackageSignature
}

// ReplaceVersionSignature replaces one signature.
func (repository *fakeSignatureRepository) ReplaceVersionSignature(
	_ context.Context,
	versionID uuid.UUID,
	signature domain.ThemePackageSignature,
) error {
	signature.VersionID = versionID
	repository.signatures[versionID] = signature
	return nil
}

// FindByVersion returns one signature.
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

// fakeSigningKeyRepository is unused by importer tests.
type fakeSigningKeyRepository struct{}

// Upsert is unused by importer tests.
func (repository fakeSigningKeyRepository) Upsert(
	context.Context,
	domain.ThemeSigningKey,
) (domain.ThemeSigningKey, error) {
	return domain.ThemeSigningKey{}, nil
}

// FindByKeyID is unused by importer tests.
func (repository fakeSigningKeyRepository) FindByKeyID(
	context.Context,
	string,
) (domain.ThemeSigningKey, error) {
	return domain.ThemeSigningKey{}, port.ErrNotFound
}

// List is unused by importer tests.
func (repository fakeSigningKeyRepository) List(context.Context) ([]domain.ThemeSigningKey, error) {
	return nil, nil
}

// fakeVerifier returns a configured signature result.
type fakeVerifier struct {
	result signing.Result
	calls  int
}

// Verify returns the configured signature result.
func (verifier *fakeVerifier) Verify(context.Context, []byte, []byte) signing.Result {
	verifier.calls++
	return verifier.result
}
