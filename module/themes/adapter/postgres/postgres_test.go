package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestThemeRepositoriesLifecycle verifies the first persistence slice works together.
func TestThemeRepositoriesLifecycle(t *testing.T) {
	repositories := newRepositories(t)
	ctx := context.Background()
	theme, err := repositories.themes.Create(ctx, testTheme())
	if err != nil {
		t.Fatalf("Create(theme) error = %v", err)
	}
	theme.Name = "RealmKit Base"
	updatedTheme, err := repositories.themes.Update(ctx, theme, theme.Version)
	if err != nil {
		t.Fatalf("Update(theme) error = %v", err)
	}
	if updatedTheme.Version != 2 || updatedTheme.Name != "RealmKit Base" {
		t.Fatalf("updatedTheme = %+v, want version 2 name", updatedTheme)
	}
	version, err := repositories.versions.Create(ctx, testVersion(theme.ID))
	if err != nil {
		t.Fatalf("Create(version) error = %v", err)
	}
	files := []domain.ThemeFile{
		testFile(version.ID, "layout/theme.liquid"),
		testFile(version.ID, "assets\\app.css"),
	}
	if err := repositories.files.ReplaceVersionFiles(ctx, version.ID, files); err != nil {
		t.Fatalf("ReplaceVersionFiles() error = %v", err)
	}
	foundFile, err := repositories.files.FindByPath(ctx, version.ID, "assets/app.css")
	if err != nil {
		t.Fatalf("FindByPath() error = %v", err)
	}
	if foundFile.Path != "assets/app.css" {
		t.Fatalf("foundFile.Path = %q, want normalized path", foundFile.Path)
	}
	storedFiles, err := repositories.files.ListByVersion(ctx, version.ID)
	if err != nil {
		t.Fatalf("ListByVersion(files) error = %v", err)
	}
	version.IntegritySHA256 = domain.CalculateVersionIntegritySHA256(integrityFiles(storedFiles))
	version.Status = domain.VersionStatusValid
	version, err = repositories.versions.Update(ctx, version, version.Version)
	if err != nil {
		t.Fatalf("Update(version) error = %v", err)
	}
	if version.IntegritySHA256 == "" || version.Version != 2 {
		t.Fatalf("version = %+v, want integrity and version 2", version)
	}
	if err := repositories.assets.ReplaceVersionAssets(ctx, version.ID, []domain.ThemeAsset{testAsset(version.ID, foundFile.ID)}); err != nil {
		t.Fatalf("ReplaceVersionAssets() error = %v", err)
	}
	if err := repositories.issues.ReplaceVersionIssues(ctx, version.ID, []domain.ThemeValidationIssue{testIssue(version.ID)}); err != nil {
		t.Fatalf("ReplaceVersionIssues() error = %v", err)
	}
	activation, err := repositories.activations.Activate(ctx, testActivation(theme.ID, version.ID))
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	current, err := repositories.activations.Current(ctx, domain.EnvironmentPublic)
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}
	if current.ID != activation.ID {
		t.Fatalf("Current() = %v, want %v", current.ID, activation.ID)
	}
	key, err := repositories.signingKeys.Upsert(ctx, testSigningKey())
	if err != nil {
		t.Fatalf("Upsert(signing key) error = %v", err)
	}
	key.Status = domain.SigningKeyRetired
	key, err = repositories.signingKeys.Upsert(ctx, key)
	if err != nil {
		t.Fatalf("Upsert(update signing key) error = %v", err)
	}
	if key.Status != domain.SigningKeyRetired {
		t.Fatalf("key.Status = %q, want retired", key.Status)
	}
	token, err := repositories.previewTokens.Create(ctx, testPreviewToken(version.ID))
	if err != nil {
		t.Fatalf("Create(preview token) error = %v", err)
	}
	if _, err := repositories.previewTokens.FindByTokenHash(ctx, token.TokenHash); err != nil {
		t.Fatalf("FindByTokenHash() error = %v", err)
	}
	if err := repositories.previewTokens.Revoke(ctx, token.ID); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if _, err := repositories.previewTokens.FindByTokenHash(ctx, token.TokenHash); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByTokenHash() after revoke error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestPublishedVersionRejectsMutation verifies published versions are immutable.
func TestPublishedVersionRejectsMutation(t *testing.T) {
	repositories := newRepositories(t)
	ctx := context.Background()
	theme, err := repositories.themes.Create(ctx, testTheme())
	if err != nil {
		t.Fatalf("Create(theme) error = %v", err)
	}
	publishedAt := time.Now().UTC()
	version := testVersion(theme.ID)
	version.Status = domain.VersionStatusPublished
	version.PublishedAt = &publishedAt
	version, err = repositories.versions.Create(ctx, version)
	if err != nil {
		t.Fatalf("Create(version) error = %v", err)
	}
	version.Label = "changed"
	if _, err := repositories.versions.Update(ctx, version, version.Version); !errors.Is(err, domain.ErrPublishedVersionImmutable) {
		t.Fatalf("Update(published) error = %v, want %v", err, domain.ErrPublishedVersionImmutable)
	}
	err = repositories.files.ReplaceVersionFiles(ctx, version.ID, []domain.ThemeFile{testFile(version.ID, "layout/theme.liquid")})
	if !errors.Is(err, domain.ErrPublishedVersionImmutable) {
		t.Fatalf("ReplaceVersionFiles(published) error = %v, want %v", err, domain.ErrPublishedVersionImmutable)
	}
}

// testRepositories groups theme repository adapters.
type testRepositories struct {
	themes        ThemeRepository
	versions      VersionRepository
	files         FileRepository
	assets        AssetRepository
	activations   ActivationRepository
	issues        ValidationIssueRepository
	signingKeys   SigningKeyRepository
	previewTokens PreviewTokenRepository
}

// newRepositories creates migrated repository adapters.
func newRepositories(t *testing.T) testRepositories {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	store := orm.NewStore(db)
	return testRepositories{
		themes:        NewThemeRepository(store),
		versions:      NewVersionRepository(store),
		files:         NewFileRepository(store),
		assets:        NewAssetRepository(store),
		activations:   NewActivationRepository(store),
		issues:        NewValidationIssueRepository(store),
		signingKeys:   NewSigningKeyRepository(store),
		previewTokens: NewPreviewTokenRepository(store),
	}
}

// testTheme returns a valid theme family.
func testTheme() domain.Theme {
	return domain.Theme{
		ID:          uuid.New(),
		Key:         "main",
		Name:        "Main",
		Description: "Default public theme.",
		Status:      domain.ThemeStatusDraft,
		Version:     1,
	}
}

// testVersion returns a valid draft version.
func testVersion(themeID uuid.UUID) domain.ThemeVersion {
	return domain.ThemeVersion{
		ID:                 uuid.New(),
		ThemeID:            themeID,
		Semver:             "1.0.0",
		Label:              "Initial",
		Status:             domain.VersionStatusDraft,
		SourceKind:         domain.SourceEditor,
		ManifestJSON:       []byte(`{"name":"Main"}`),
		SettingsSchemaJSON: []byte(`{}`),
		SettingsDataJSON:   []byte(`{}`),
		Version:            1,
	}
}

// testFile returns a valid version file.
func testFile(versionID uuid.UUID, path domain.FilePath) domain.ThemeFile {
	return domain.ThemeFile{
		ID:             uuid.New(),
		VersionID:      versionID,
		Kind:           domain.FileKindLayout,
		Path:           path,
		ContentSHA256:  domain.Digest(uuid.NewString()),
		ContentStorage: "themes/main/" + uuid.NewString(),
		ContentText:    "content",
		SizeBytes:      7,
	}
}

// testAsset returns a valid theme asset.
func testAsset(versionID uuid.UUID, fileID uuid.UUID) domain.ThemeAsset {
	return domain.ThemeAsset{
		ID:             uuid.New(),
		VersionID:      versionID,
		FileID:         fileID,
		Path:           "assets/app.css",
		ContentType:    "text/css",
		SizeBytes:      12,
		ContentSHA256:  "asset-sha",
		StorageKey:     "themes/main/assets/app.css",
		PublicURL:      "/theme-assets/app.css",
		IntegrityValue: "sha256-asset",
	}
}

// testIssue returns a valid validation issue.
func testIssue(versionID uuid.UUID) domain.ThemeValidationIssue {
	return domain.ThemeValidationIssue{
		ID:        uuid.New(),
		VersionID: versionID,
		Severity:  domain.SeverityWarning,
		Code:      domain.IssueUnsafeCSS,
		Path:      "assets/app.css",
		Message:   "External import ignored.",
		Line:      1,
		Column:    1,
		Details:   []byte(`{"rule":"external-import"}`),
	}
}

// testActivation returns a public activation.
func testActivation(themeID uuid.UUID, versionID uuid.UUID) domain.ThemeActivation {
	return domain.ThemeActivation{
		ID:          uuid.New(),
		ThemeID:     themeID,
		VersionID:   versionID,
		Environment: domain.EnvironmentPublic,
		Reason:      "Initial publication",
	}
}

// testSigningKey returns a trusted signing key.
func testSigningKey() domain.ThemeSigningKey {
	return domain.ThemeSigningKey{
		ID:          uuid.New(),
		KeyID:       "realmkit:test",
		Algorithm:   domain.SignatureAlgorithmEd25519,
		PublicKey:   "public-key",
		TrustLevel:  domain.TrustLevelOperator,
		Status:      domain.SigningKeyTrusted,
		Source:      domain.SigningKeySourceEnvironment,
		Description: "Test key",
	}
}

// testPreviewToken returns a preview token.
func testPreviewToken(versionID uuid.UUID) domain.ThemePreviewToken {
	return domain.ThemePreviewToken{
		ID:            uuid.New(),
		VersionID:     versionID,
		TokenHash:     "preview-token-hash",
		PersonaKind:   domain.PersonaModerator,
		PersonaSource: domain.PersonaSourceSynthetic,
		ExpiresAt:     time.Now().UTC().Add(time.Hour),
	}
}

// integrityFiles converts stored files into integrity hash inputs.
func integrityFiles(files []domain.ThemeFile) []domain.IntegrityFile {
	inputs := make([]domain.IntegrityFile, 0, len(files))
	for _, file := range files {
		inputs = append(inputs, domain.IntegrityFile{Path: file.Path, ContentSHA256: file.ContentSHA256})
	}
	return inputs
}
