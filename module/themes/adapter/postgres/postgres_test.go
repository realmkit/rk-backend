package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
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
	foundBySource, err := repositories.versions.FindBySourceReference(ctx, theme.ID, version.SourceReference)
	if err != nil {
		t.Fatalf("FindBySourceReference() error = %v", err)
	}
	if foundBySource.ID != version.ID {
		t.Fatalf("FindBySourceReference() = %v, want %v", foundBySource.ID, version.ID)
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
	if err := repositories.signatures.ReplaceVersionSignature(ctx, version.ID, testSignature(version.ID)); err != nil {
		t.Fatalf("ReplaceVersionSignature() error = %v", err)
	}
	signature, err := repositories.signatures.FindByVersion(ctx, version.ID)
	if err != nil {
		t.Fatalf("FindByVersion(signature) error = %v", err)
	}
	if signature.VerificationStatus != domain.SignatureVerified {
		t.Fatalf("signature status = %q, want verified", signature.VerificationStatus)
	}
	activation, err := repositories.activations.Activate(ctx, testActivation(theme.ID, version.ID))
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	foundActivation, err := repositories.activations.FindByID(ctx, activation.ID)
	if err != nil {
		t.Fatalf("FindByID(activation) error = %v", err)
	}
	if string(foundActivation.SettingsDataJSON) != `{"accent":"lime"}` {
		t.Fatalf("SettingsDataJSON = %s, want activation settings", foundActivation.SettingsDataJSON)
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
