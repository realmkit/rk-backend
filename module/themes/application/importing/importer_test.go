package importing

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/application/signing"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// TestServiceImportsSafePackagePersistsDraftAssets verifies successful zip import.
func TestServiceImportsSafePackagePersistsDraftAssets(t *testing.T) {
	repositories := newFakeRepositories()
	store := &fakeStore{objects: map[string][]byte{}}
	service := NewService(repositories.ports(), store, Config{}, nil)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter","version":"1.0.0"}`),
		"layout/theme.liquid": []byte(`<main>{{ content_for_layout }}</main>`),
		"assets/logo.png":     {0x89, 0x50, 0x4e, 0x47},
	})
	result, err := service.Import(context.Background(), Command{
		ThemeID:          uuid.New(),
		IdempotencyKey:   "safe-import",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !result.Imported || result.Version.Status != domain.VersionStatusDraft {
		t.Fatalf("result = %+v, want imported draft", result)
	}
	if len(repositories.files.files[result.Version.ID]) != 3 {
		t.Fatalf("files = %d, want 3", len(repositories.files.files[result.Version.ID]))
	}
	if len(repositories.assets.assets[result.Version.ID]) != 1 {
		t.Fatalf("assets = %d, want 1", len(repositories.assets.assets[result.Version.ID]))
	}
	if len(store.objects) != 1 {
		t.Fatalf("stored objects = %d, want 1", len(store.objects))
	}
}

// TestServiceReusesIdempotentImport verifies source reference retry behavior.
func TestServiceReusesIdempotentImport(t *testing.T) {
	repositories := newFakeRepositories()
	service := NewService(repositories.ports(), &fakeStore{objects: map[string][]byte{}}, Config{}, nil)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter","version":"1.0.0"}`),
	})
	themeID := uuid.New()
	first, err := service.Import(context.Background(), Command{
		ThemeID:          themeID,
		IdempotencyKey:   "retry",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("first Import() error = %v", err)
	}
	second, err := service.Import(context.Background(), Command{ThemeID: themeID, IdempotencyKey: "retry"})
	if err != nil {
		t.Fatalf("second Import() error = %v", err)
	}
	if !second.Reused || second.Version.ID != first.Version.ID {
		t.Fatalf("second = %+v, want reused first version", second)
	}
}

// TestServiceRejectsUnsafePackageIssues verifies blocking package diagnostics.
func TestServiceRejectsUnsafePackageIssues(t *testing.T) {
	repositories := newFakeRepositories()
	service := NewService(repositories.ports(), &fakeStore{objects: map[string][]byte{}}, Config{}, nil)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json":        []byte(`{"name":"Starter","version":"1.0.0"}`),
		"layout/theme.liquid":        []byte(`layout`),
		"layout/./theme.liquid":      []byte(`duplicate`),
		"templates/index.liquid":     {0xff, 0xfe},
		"../templates/escape.liquid": []byte(`escape`),
	})
	result, err := service.Import(context.Background(), Command{
		ThemeID:          uuid.New(),
		IdempotencyKey:   "unsafe",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	codes := issueCodes(result.Issues)
	assertIssue(t, codes, domain.IssueUnsafePath)
	assertIssue(t, codes, domain.IssueDuplicatePath)
	assertIssue(t, codes, domain.IssueInvalidUTF8)
	if result.Version.Status != domain.VersionStatusInvalid {
		t.Fatalf("Status = %q, want invalid", result.Version.Status)
	}
	if len(repositories.files.files[result.Version.ID]) != 0 {
		t.Fatalf("files persisted = %d, want 0", len(repositories.files.files[result.Version.ID]))
	}
}

// TestExtractPackageEnforcesLimits verifies size, count, and compression diagnostics.
func TestExtractPackageEnforcesLimits(t *testing.T) {
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter"}`),
		"layout/theme.liquid": bytes.Repeat([]byte("a"), 128),
	})
	_, issues, err := extractPackage(reader, size, Config{
		MaxPackageBytes:     DefaultMaxPackageBytes,
		MaxExtractedBytes:   DefaultMaxExtractedBytes,
		MaxFileCount:        1,
		MaxTextFileBytes:    8,
		MaxCompressionRatio: 0.1,
	}.Defaults())
	if err != nil {
		t.Fatalf("extractPackage() error = %v", err)
	}
	codes := issueCodes(issues)
	assertIssue(t, codes, domain.IssueFileCountTooLarge)
	assertIssue(t, codes, domain.IssueTextFileTooLarge)
	assertIssue(t, codes, domain.IssueCompressionRatioTooLarge)
	_, packageIssues, err := extractPackage(bytes.NewReader([]byte("too large")), 99, Config{MaxPackageBytes: 4}.Defaults())
	if err != nil {
		t.Fatalf("extractPackage(large) error = %v", err)
	}
	assertIssue(t, issueCodes(packageIssues), domain.IssuePackageTooLarge)
}

// TestServicePersistsSignatureResult verifies verifier output persistence.
func TestServicePersistsSignatureResult(t *testing.T) {
	repositories := newFakeRepositories()
	verifier := &fakeVerifier{result: signing.Result{
		Signature: domain.ThemePackageSignature{
			ID:                 uuid.New(),
			Algorithm:          domain.SignatureAlgorithmEd25519,
			VerificationStatus: domain.SignatureVerified,
		},
	}}
	service := NewService(repositories.ports(), &fakeStore{objects: map[string][]byte{}}, Config{}, verifier)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter"}`),
	})
	result, err := service.Import(context.Background(), Command{
		ThemeID:          uuid.New(),
		IdempotencyKey:   "signed",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	signature := repositories.signatures.signatures[result.Version.ID]
	if signature.VerificationStatus != domain.SignatureVerified {
		t.Fatalf("signature = %+v, want verified", signature)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
}

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

// fakeStore stores objects in memory.
type fakeStore struct {
	objects map[string][]byte
}

// Health reports fake storage health.
func (store *fakeStore) Health(context.Context) error {
	return nil
}

// Put stores object bytes.
func (store *fakeStore) Put(_ context.Context, object storage.Object, body io.Reader) (storage.StoredObject, error) {
	content, err := io.ReadAll(body)
	if err != nil {
		return storage.StoredObject{}, err
	}
	store.objects[object.Key] = content
	return storage.StoredObject{Key: object.Key}, nil
}

// Delete deletes one object.
func (store *fakeStore) Delete(_ context.Context, key string) error {
	delete(store.objects, key)
	return nil
}

// PresignPut is unused by importer tests.
func (store *fakeStore) PresignPut(context.Context, storage.PresignPutRequest) (storage.PresignedRequest, error) {
	return storage.PresignedRequest{}, errors.New("unused")
}

// PresignGet is unused by importer tests.
func (store *fakeStore) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "", errors.New("unused")
}

// Head is unused by importer tests.
func (store *fakeStore) Head(context.Context, string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{}, errors.New("unused")
}

// zipPackage builds a test zip package.
func zipPackage(t *testing.T, files map[string][]byte) (io.Reader, int64) {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return bytes.NewReader(buffer.Bytes()), int64(buffer.Len())
}

// issueCodes returns a set of issue codes.
func issueCodes(issues []domain.ThemeValidationIssue) map[domain.ValidationIssueCode]struct{} {
	codes := map[domain.ValidationIssueCode]struct{}{}
	for _, issue := range issues {
		codes[issue.Code] = struct{}{}
	}
	return codes
}

// assertIssue verifies an issue code exists.
func assertIssue(
	t *testing.T,
	codes map[domain.ValidationIssueCode]struct{},
	code domain.ValidationIssueCode,
) {
	t.Helper()
	if _, ok := codes[code]; !ok {
		t.Fatalf("issue %q missing from %v", code, codes)
	}
}
