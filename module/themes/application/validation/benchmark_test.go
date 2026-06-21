package validation

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// BenchmarkValidateThemePackage measures full static validation allocation cost.
func BenchmarkValidateThemePackage(b *testing.B) {
	files := benchmarkValidationFiles()
	repositories := validationRepositories(files)
	service := NewService(repositories)
	command := Command{VersionID: repositories.Versions.(*fakeVersionRepository).version.ID}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Validate(context.Background(), command); err != nil {
			b.Fatalf("Validate() error = %v", err)
		}
	}
}

// BenchmarkValidateLiquidFiles measures dependency and RealmKit tag scanning.
func BenchmarkValidateLiquidFiles(b *testing.B) {
	files := attachVersion(uuid.New(), benchmarkValidationFiles())
	index := indexFiles(files)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		issues := validateLiquidFiles(context.Background(), files, index)
		if len(issues) != 0 {
			b.Fatalf("validateLiquidFiles() issues = %d, want 0", len(issues))
		}
	}
}

// benchmarkValidationFiles returns a realistic valid theme package.
func benchmarkValidationFiles() []domain.ThemeFile {
	files := []domain.ThemeFile{
		testFile("layout/theme.liquid", domain.FileKindLayout, `{% rk_section "hero" %}{% rk_section "stats" %}`),
		testFile("sections/hero.liquid", domain.FileKindSection, benchmarkLiquid("hero")),
		testFile("sections/stats.liquid", domain.FileKindSection, benchmarkLiquid("stats")),
		testFile("snippets/card.liquid", domain.FileKindSnippet, `<article>{{ card.title }}</article>`),
		testFile("assets/theme.css", domain.FileKindAsset, `.theme{display:grid}`),
		testFile("assets/app.js", domain.FileKindAsset, `console.log("realmkit")`),
		testFile("config/settings_schema.json", domain.FileKindConfig, `{"required":["brand"]}`),
		testFile("config/settings_data.json", domain.FileKindConfig, `{"brand":"RealmKit"}`),
	}
	for _, route := range domain.RouteKinds() {
		files = append(files, testFile(routeTemplatePath(route), domain.FileKindTemplate, benchmarkLiquid(string(route))))
	}
	for index := 0; index < 24; index++ {
		files = append(
			files,
			testFile(
				domain.FilePath(fmt.Sprintf("snippets/item_%02d.liquid", index)),
				domain.FileKindSnippet,
				`<span>{{ item.name }}</span>`,
			),
		)
	}
	return files
}

// benchmarkLiquid returns dependency-heavy but valid Liquid content.
func benchmarkLiquid(name string) string {
	var builder strings.Builder
	builder.WriteString(`<main>`)
	builder.WriteString(name)
	builder.WriteString(`{% render "card" %}`)
	builder.WriteString(`{{ "theme.css" | asset_url }}`)
	builder.WriteString(`{{ "app.js" | asset_url }}`)
	builder.WriteString(`</main>`)
	return builder.String()
}
