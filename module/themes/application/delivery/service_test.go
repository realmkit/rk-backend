package delivery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// TestDeliveryReturnsActivationManifestFilesAssetsAndReports verifies render delivery data.
func TestDeliveryReturnsActivationManifestFilesAssetsAndReports(t *testing.T) {
	repositories, themeID, versionID := deliveryRepositories()
	service := NewService(repositories, fixedClock())
	activation, err := service.ActiveActivation(context.Background(), domain.EnvironmentPublic)
	if err != nil {
		t.Fatalf("ActiveActivation() error = %v", err)
	}
	if activation.Cache.CacheControl != revalidateCacheControl {
		t.Fatalf("activation cache = %q, want revalidate", activation.Cache.CacheControl)
	}
	manifest, err := service.Manifest(context.Background(), themeID, versionID)
	if err != nil {
		t.Fatalf("Manifest() error = %v", err)
	}
	if manifest.Cache.CacheControl != immutableCacheControl || manifest.CSP["script_src"] == nil {
		t.Fatalf("manifest = %+v, want immutable cache and CSP", manifest)
	}
	file, err := service.File(context.Background(), versionID, "layout/theme.liquid")
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if file.Cache.ETag != `"layout-sha"` {
		t.Fatalf("file ETag = %q, want content hash", file.Cache.ETag)
	}
	asset, err := service.Asset(context.Background(), versionID, "assets/app.js")
	if err != nil {
		t.Fatalf("Asset() error = %v", err)
	}
	if asset.Cache.CacheControl != immutableCacheControl {
		t.Fatalf("asset cache = %q, want immutable", asset.Cache.CacheControl)
	}
	report, err := service.ValidationReport(context.Background(), versionID)
	if err != nil {
		t.Fatalf("ValidationReport() error = %v", err)
	}
	if len(report.Issues) != 1 {
		t.Fatalf("issues = %d, want 1", len(report.Issues))
	}
}

// TestPreviewTokenLifecycle verifies token hashing, defaults, and expiry.
func TestPreviewTokenLifecycle(t *testing.T) {
	repositories, _, versionID := deliveryRepositories()
	service := NewService(repositories, fixedClock())
	created, err := service.CreatePreviewToken(context.Background(), CreatePreviewTokenCommand{VersionID: versionID})
	if err != nil {
		t.Fatalf("CreatePreviewToken() error = %v", err)
	}
	if created.Token == "" || created.Preview.TokenHash == created.Token {
		t.Fatalf("created = %+v, want raw token and stored hash", created)
	}
	valid, err := service.ValidatePreviewToken(context.Background(), ValidatePreviewTokenCommand{Token: created.Token})
	if err != nil {
		t.Fatalf("ValidatePreviewToken() error = %v", err)
	}
	if valid.PersonaKind != domain.PersonaAnonymous {
		t.Fatalf("PersonaKind = %q, want anonymous", valid.PersonaKind)
	}
	expired := NewService(repositories, func() time.Time { return fixedClock()().Add(time.Hour) })
	if _, err := expired.ValidatePreviewToken(context.Background(), ValidatePreviewTokenCommand{Token: created.Token}); !errors.Is(err, port.ErrInvalidState) {
		t.Fatalf("expired error = %v, want invalid state", err)
	}
}

// TestSanitizerProfiles verifies reusable rich-text policies.
func TestSanitizerProfiles(t *testing.T) {
	profile, err := SanitizerProfileFor(domain.ProfileForumPost)
	if err != nil {
		t.Fatalf("SanitizerProfileFor() error = %v", err)
	}
	if !profile.AllowsElement("blockquote") || profile.AllowsElement("script") {
		t.Fatalf("profile = %+v, want safe forum allowlist", profile)
	}
	if _, err := SanitizerProfileFor("missing"); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("missing profile error = %v, want not found", err)
	}
}

// deliveryRepositories returns a complete delivery fixture.
func deliveryRepositories() (Repositories, uuid.UUID, uuid.UUID) {
	themeID := uuid.New()
	versionID := uuid.New()
	theme := domain.Theme{ID: themeID, Key: "main", Name: "Main", Status: domain.ThemeStatusAvailable}
	version := domain.ThemeVersion{
		ID: versionID, ThemeID: themeID, Status: domain.VersionStatusPublished, IntegritySHA256: "version-sha",
		ManifestJSON:       []byte(`{"route_coverage":{"home":true},"dependency_graph":{"sections":["hero"]}}`),
		SettingsSchemaJSON: []byte(`{"required":["brand"]}`), SettingsDataJSON: []byte(`{"brand":"RealmKit"}`),
	}
	return Repositories{
		Themes:      fakeThemeRepository{themes: map[uuid.UUID]domain.Theme{themeID: theme}},
		Versions:    fakeVersionRepository{versions: map[uuid.UUID]domain.ThemeVersion{versionID: version}},
		Files:       fakeFileRepository{files: map[uuid.UUID][]domain.ThemeFile{versionID: deliveryFiles(versionID)}},
		Assets:      fakeAssetRepository{assets: map[uuid.UUID][]domain.ThemeAsset{versionID: deliveryAssets(versionID)}},
		Activations: fakeActivationRepository{current: domain.ThemeActivation{ID: uuid.New(), ThemeID: themeID, VersionID: versionID, Environment: domain.EnvironmentPublic, SettingsDataJSON: []byte(`{"brand":"RealmKit"}`), ActivatedAt: fixedClock()()}},
		Issues:      fakeIssueRepository{issues: map[uuid.UUID][]domain.ThemeValidationIssue{versionID: {{ID: uuid.New(), VersionID: versionID, Severity: domain.SeverityWarning, Code: domain.IssueUnsafeCSS}}}},
		PreviewTokens: fakePreviewTokenRepository{
			tokens: map[string]domain.ThemePreviewToken{},
		},
	}, themeID, versionID
}

// deliveryFiles returns source files for delivery tests.
func deliveryFiles(versionID uuid.UUID) []domain.ThemeFile {
	return []domain.ThemeFile{{ID: uuid.New(), VersionID: versionID, Kind: domain.FileKindLayout, Path: "layout/theme.liquid", ContentSHA256: "layout-sha", ContentText: "layout"}}
}

// deliveryAssets returns asset files for delivery tests.
func deliveryAssets(versionID uuid.UUID) []domain.ThemeAsset {
	return []domain.ThemeAsset{{ID: uuid.New(), VersionID: versionID, Path: "assets/app.js", ContentType: "application/javascript", ContentSHA256: "asset-sha", PublicURL: "/theme-assets/app.js", IntegrityValue: "sha256-asset"}}
}

// fixedClock returns a deterministic delivery clock.
func fixedClock() Clock {
	return func() time.Time { return time.Date(2026, time.June, 19, 12, 0, 0, 0, time.UTC) }
}
