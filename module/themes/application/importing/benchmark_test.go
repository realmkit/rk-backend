package importing

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// benchmarkPackageFiles stores extracted package files.
var benchmarkPackageFiles []packageFile

// benchmarkPackageIssues stores extraction diagnostics.
var benchmarkPackageIssues []domain.ThemeValidationIssue

// BenchmarkExtractPackage measures zip package intake, normalization, digesting, and text classification.
func BenchmarkExtractPackage(b *testing.B) {
	archiveBytes, packageSize := benchmarkZipBytes(b)
	cfg := Config{}.Defaults()
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		files, issues, err := extractPackage(ctx, bytes.NewReader(archiveBytes), packageSize, cfg)
		if err != nil {
			b.Fatalf("extractPackage() error = %v", err)
		}
		benchmarkPackageFiles = files
		benchmarkPackageIssues = issues
	}
}

// benchmarkZipBytes returns a representative theme package archive.
func benchmarkZipBytes(b *testing.B) ([]byte, int64) {
	b.Helper()
	files := []struct {
		name    string
		content []byte
	}{
		{name: "realmkit-theme.json", content: []byte(`{"name":"Benchmark","version":"1.0.0"}`)},
		{name: "layout/theme.liquid", content: []byte(`<main>{{ content_for_layout }}</main>`)},
		{name: "templates/home.liquid", content: []byte(`<section>{% section "hero" %}</section>`)},
		{name: "sections/hero.liquid", content: []byte(`<rk:slot name="title">Benchmark</rk:slot>`)},
		{name: "snippets/card.liquid", content: []byte(`<article>{{ title }}</article>`)},
		{name: "assets/app.js", content: []byte(`console.log("realmkit")`)},
		{name: "assets/app.css", content: []byte(`.theme{display:block}`)},
		{name: "assets/logo.png", content: []byte{0x89, 0x50, 0x4e, 0x47}},
		{name: "config/settings_schema.json", content: []byte(`{"required":["brand"]}`)},
		{name: "locales/en.json", content: []byte(`{"hello":"Hello"}`)},
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, file := range files {
		entry, err := writer.Create(file.name)
		if err != nil {
			b.Fatalf("Create(%q) error = %v", file.name, err)
		}
		if _, err := entry.Write(file.content); err != nil {
			b.Fatalf("Write(%q) error = %v", file.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		b.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes(), int64(buffer.Len())
}
