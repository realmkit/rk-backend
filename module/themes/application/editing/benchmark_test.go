package editing

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// benchmarkEditedFile stores editing helper benchmark output.
var benchmarkEditedFile domain.ThemeFile

// BenchmarkUpdateFile measures editor file replacement, digesting, and path normalization.
func BenchmarkUpdateFile(b *testing.B) {
	versionID := uuid.New()
	fileID := uuid.New()
	files := []domain.ThemeFile{{
		ID:            fileID,
		VersionID:     versionID,
		Kind:          domain.FileKindTemplate,
		Path:          "templates/home.liquid",
		ContentText:   "home",
		ContentSHA256: digest("home"),
		SizeBytes:     4,
	}}
	command := WriteFileCommand{
		VersionID:   versionID,
		FileID:      fileID,
		Path:        "templates/home.liquid",
		Kind:        domain.FileKindTemplate,
		ContentText: benchmarkThemeContent,
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		file, err := updateFile(files, command)
		if err != nil {
			b.Fatalf("updateFile() error = %v", err)
		}
		benchmarkEditedFile = file
	}
}

// benchmarkThemeContent is representative editable template content.
const benchmarkThemeContent = `<main><rk:slot name="content">{{ section.settings.title }}</rk:slot></main>`
