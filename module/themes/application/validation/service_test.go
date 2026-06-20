package validation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// TestValidateReportsStaticIssues verifies validation diagnostics.
func TestValidateReportsStaticIssues(t *testing.T) {
	repositories := validationRepositories([]domain.ThemeFile{
		testFile("layout/theme.liquid", domain.FileKindLayout, `{% rk_magic %}{{ broken`),
		testFile("templates/home.liquid", domain.FileKindTemplate, `{% rk_section "hero" %}{{ "remote.css" | asset_url }}`),
		testFile("assets/theme.css", domain.FileKindAsset, `@import url("https://example.test/x.css");`),
		testFile("assets/app.js", domain.FileKindAsset, `eval("bad")`),
		testFile("locales/en.json", domain.FileKindLocale, `{`),
		testFile("config/settings_schema.json", domain.FileKindConfig, `{"required":["brand"]}`),
		testFile("config/settings_data.json", domain.FileKindConfig, `{}`),
	})
	service := NewService(repositories)
	result, err := service.Validate(context.Background(), Command{VersionID: repositories.Versions.(*fakeVersionRepository).version.ID})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	codes := validationIssueCodes(result.Issues)
	for _, code := range []domain.ValidationIssueCode{
		domain.IssueUnknownRealmKitTag,
		domain.IssueInvalidLiquid,
		domain.IssueMissingDependency,
		domain.IssueUnsafeCSS,
		domain.IssueUnsafeJavaScript,
		domain.IssueInvalidLocale,
		domain.IssueInvalidSettingsData,
		domain.IssueMissingRequiredTemplate,
	} {
		if _, ok := codes[code]; !ok {
			t.Fatalf("issue %q missing from %v", code, codes)
		}
	}
	if result.Version.Status != domain.VersionStatusInvalid {
		t.Fatalf("Status = %q, want invalid", result.Version.Status)
	}
}

// TestValidateMarksValidAndBuildsManifest verifies successful validation.
func TestValidateMarksValidAndBuildsManifest(t *testing.T) {
	files := []domain.ThemeFile{
		testFile("layout/theme.liquid", domain.FileKindLayout, `{% rk_section "hero" %}`),
		testFile("sections/hero.liquid", domain.FileKindSection, `<section>Hero</section>`),
		testFile("config/settings_schema.json", domain.FileKindConfig, `{"required":["brand"]}`),
		testFile("config/settings_data.json", domain.FileKindConfig, `{"brand":"RealmKit"}`),
	}
	for _, route := range domain.RouteKinds() {
		files = append(files, testFile(routeTemplatePath(route), domain.FileKindTemplate, `<main>{{ page.title }}</main>`))
	}
	repositories := validationRepositories(files)
	service := NewService(repositories)
	result, err := service.Validate(context.Background(), Command{VersionID: repositories.Versions.(*fakeVersionRepository).version.ID})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Version.Status != domain.VersionStatusValid {
		t.Fatalf("Status = %q, want valid", result.Version.Status)
	}
	var manifest map[string]any
	if err := json.Unmarshal(result.ManifestJSON, &manifest); err != nil {
		t.Fatalf("manifest JSON error = %v", err)
	}
	if manifest["route_coverage"] == nil || manifest["dependency_graph"] == nil {
		t.Fatalf("manifest = %v, want coverage and dependencies", manifest)
	}
}

// TestValidateHonorsCancellation verifies validation stops when context is done.
func TestValidateHonorsCancellation(t *testing.T) {
	repositories := validationRepositories([]domain.ThemeFile{
		testFile("layout/theme.liquid", domain.FileKindLayout, `<main></main>`),
	})
	service := NewService(repositories)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Validate(ctx, Command{VersionID: repositories.Versions.(*fakeVersionRepository).version.ID})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Validate() error = %v, want canceled", err)
	}
}

// validationRepositories returns fake validation repositories.
func validationRepositories(files []domain.ThemeFile) Repositories {
	versionID := uuid.New()
	return Repositories{
		Versions: &fakeVersionRepository{version: domain.ThemeVersion{
			ID:                 versionID,
			ThemeID:            uuid.New(),
			Status:             domain.VersionStatusDraft,
			ManifestJSON:       []byte(`{}`),
			SettingsSchemaJSON: []byte(`{}`),
			SettingsDataJSON:   []byte(`{}`),
			Version:            1,
		}},
		Files:  &fakeFileRepository{files: attachVersion(versionID, files)},
		Issues: &fakeIssueRepository{},
	}
}

// testFile returns one validation file.
func testFile(filePath domain.FilePath, kind domain.FileKind, content string) domain.ThemeFile {
	return domain.ThemeFile{ID: uuid.New(), Kind: kind, Path: filePath, ContentText: content, ContentSHA256: "sha", SizeBytes: int64(len(content))}
}

// attachVersion sets a version ID on files.
func attachVersion(versionID uuid.UUID, files []domain.ThemeFile) []domain.ThemeFile {
	for index := range files {
		files[index].VersionID = versionID
	}
	return files
}

// validationIssueCodes returns issue codes as a set.
func validationIssueCodes(issues []domain.ThemeValidationIssue) map[domain.ValidationIssueCode]struct{} {
	codes := map[domain.ValidationIssueCode]struct{}{}
	for _, issue := range issues {
		codes[issue.Code] = struct{}{}
	}
	return codes
}
