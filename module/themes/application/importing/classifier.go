package importing

import (
	"mime"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// packageFile is one safe normalized package file.
type packageFile struct {
	path        domain.FilePath
	kind        domain.FileKind
	bytes       []byte
	sha256      domain.Digest
	contentType string
	text        bool
}

// normalizePackagePath returns a safe POSIX package path.
func normalizePackagePath(raw string) (domain.FilePath, bool) {
	replaced := strings.ReplaceAll(raw, "\\", "/")
	if raw == "" || !utf8.ValidString(raw) || path.IsAbs(replaced) || filepath.IsAbs(raw) {
		return "", false
	}
	cleaned := path.Clean(replaced)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false
	}
	return domain.NormalizeFilePath(domain.FilePath(cleaned)), true
}

// classifyFileKind returns the logical kind for a package path.
func classifyFileKind(filePath domain.FilePath) domain.FileKind {
	value := string(filePath)
	switch {
	case value == "realmkit-theme.json":
		return domain.FileKindManifest
	case value == "realmkit-theme.sig.json":
		return domain.FileKindSignature
	case strings.HasPrefix(value, "layout/"):
		return domain.FileKindLayout
	case strings.HasPrefix(value, "templates/"):
		return domain.FileKindTemplate
	case strings.HasPrefix(value, "sections/"):
		return domain.FileKindSection
	case strings.HasPrefix(value, "snippets/"):
		return domain.FileKindSnippet
	case strings.HasPrefix(value, "assets/"):
		return domain.FileKindAsset
	case strings.HasPrefix(value, "config/"):
		return domain.FileKindConfig
	case strings.HasPrefix(value, "locales/"):
		return domain.FileKindLocale
	default:
		return domain.FileKindAsset
	}
}

// isTextFile reports whether content should be stored as editor text.
func isTextFile(filePath domain.FilePath, kind domain.FileKind, content []byte) bool {
	if !utf8.Valid(content) {
		return false
	}
	return expectsTextFile(filePath, kind)
}

// expectsTextFile reports whether a package file is editor text by contract.
func expectsTextFile(filePath domain.FilePath, kind domain.FileKind) bool {
	if kind != domain.FileKindAsset {
		return true
	}
	switch strings.ToLower(path.Ext(string(filePath))) {
	case ".css", ".js", ".json", ".liquid", ".html", ".svg", ".txt", ".xml", ".map":
		return true
	default:
		return false
	}
}

// contentTypeForPath returns an HTTP content type for a package file.
func contentTypeForPath(filePath domain.FilePath) string {
	if value := mime.TypeByExtension(path.Ext(string(filePath))); value != "" {
		return value
	}
	return "application/octet-stream"
}
