package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testRepositories groups theme repository adapters.
type testRepositories struct {
	themes        ThemeRepository
	versions      VersionRepository
	files         FileRepository
	assets        AssetRepository
	activations   ActivationRepository
	issues        ValidationIssueRepository
	signatures    SignatureRepository
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
		signatures:    NewSignatureRepository(store),
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
		SourceReference:    "editor:test",
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

// testSignature returns package signature data.
func testSignature(versionID uuid.UUID) domain.ThemePackageSignature {
	now := time.Now().UTC()
	return domain.ThemePackageSignature{
		ID:                 uuid.New(),
		VersionID:          versionID,
		KeyID:              "realmkit:test",
		Algorithm:          domain.SignatureAlgorithmEd25519,
		VerificationStatus: domain.SignatureVerified,
		Signature:          "signature",
		SignedManifestHash: "manifest-sha",
		VerifiedAt:         &now,
	}
}
