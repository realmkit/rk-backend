package importing

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// manifestDocument contains importer-facing manifest metadata.
type manifestDocument struct {
	Name    string `json:"name"`    // Name stores the name value.
	Version string `json:"version"` // Version stores the version value.
}

// manifestPayload contains decoded manifest import data.
type manifestPayload struct {
	raw      []byte           // raw stores the raw value.
	document manifestDocument // document stores the document value.
	files    []packageFile    // files stores the files value.
}

// extractPackage reads a bounded zip package into normalized files.
func extractPackage(ctx context.Context, reader io.Reader, packageSize int64, cfg Config) ([]packageFile, []domain.ThemeValidationIssue, error) {
	if err := checkContext(ctx); err != nil {
		return nil, nil, err
	}
	if reader == nil {
		return nil, []domain.ThemeValidationIssue{newIssue(domain.IssuePackageTooLarge, "", "Package body is required.")}, nil
	}
	if packageSize > cfg.MaxPackageBytes {
		return nil, []domain.ThemeValidationIssue{newIssue(domain.IssuePackageTooLarge, "", "Theme package is larger than allowed.")}, nil
	}
	data, err := io.ReadAll(io.LimitReader(reader, cfg.MaxPackageBytes+1))
	if err != nil {
		return nil, nil, err
	}
	if err := checkContext(ctx); err != nil {
		return nil, nil, err
	}
	if int64(len(data)) > cfg.MaxPackageBytes {
		return nil, []domain.ThemeValidationIssue{newIssue(domain.IssuePackageTooLarge, "", "Theme package is larger than allowed.")}, nil
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, []domain.ThemeValidationIssue{newIssue(domain.IssueInvalidManifest, "", "Theme package must be a valid zip archive.")}, nil
	}
	return extractArchiveFiles(ctx, archive, cfg)
}

// extractArchiveFiles reads normalized files from a zip archive.
func extractArchiveFiles(ctx context.Context, archive *zip.Reader, cfg Config) ([]packageFile, []domain.ThemeValidationIssue, error) {
	seen := map[domain.FilePath]struct{}{}
	files := make([]packageFile, 0, len(archive.File))
	issues := make([]domain.ThemeValidationIssue, 0)
	var extracted int64
	for _, entry := range archive.File {
		if err := checkContext(ctx); err != nil {
			return nil, nil, err
		}
		if entry.FileInfo().IsDir() {
			continue
		}
		if len(files) >= cfg.MaxFileCount {
			issues = append(issues, newIssue(domain.IssueFileCountTooLarge, "", "Theme package contains too many files."))
			continue
		}
		file, fileIssues, err := extractArchiveFile(ctx, entry, cfg, seen)
		issues = append(issues, fileIssues...)
		if err != nil {
			return nil, nil, err
		}
		if file.path == "" {
			continue
		}
		extracted += int64(len(file.bytes))
		if extracted > cfg.MaxExtractedBytes {
			issues = append(issues, newIssue(domain.IssueExtractedPackageTooLarge, file.path, "Theme package expands larger than allowed."))
			continue
		}
		files = append(files, file)
	}
	return files, issues, nil
}

// extractArchiveFile reads one zip entry.
func extractArchiveFile(
	ctx context.Context,
	entry *zip.File,
	cfg Config,
	seen map[domain.FilePath]struct{},
) (packageFile, []domain.ThemeValidationIssue, error) {
	if err := checkContext(ctx); err != nil {
		return packageFile{}, nil, err
	}
	issues := make([]domain.ThemeValidationIssue, 0)
	filePath, ok := normalizePackagePath(entry.Name)
	if !ok {
		return packageFile{}, []domain.ThemeValidationIssue{newIssue(domain.IssueUnsafePath, "", "Theme package contains an unsafe path.")}, nil
	}
	if _, exists := seen[filePath]; exists {
		return packageFile{}, []domain.ThemeValidationIssue{newIssue(domain.IssueDuplicatePath, filePath, "Theme package contains duplicate paths.")}, nil
	}
	seen[filePath] = struct{}{}
	if ratioTooHigh(entry, cfg.MaxCompressionRatio) {
		issues = append(issues, newIssue(domain.IssueCompressionRatioTooLarge, filePath, "Theme package compression ratio is suspicious."))
	}
	content, err := readZipEntry(ctx, entry, cfg.MaxExtractedBytes)
	if err != nil {
		return packageFile{}, nil, err
	}
	kind := classifyFileKind(filePath)
	text := isTextFile(filePath, kind, content)
	if expectsTextFile(filePath, kind) && !utf8.Valid(content) {
		issues = append(issues, newIssue(domain.IssueInvalidUTF8, filePath, "Theme text file is not valid UTF-8."))
	}
	if text && int64(len(content)) > cfg.MaxTextFileBytes {
		issues = append(issues, newIssue(domain.IssueTextFileTooLarge, filePath, "Theme text file is larger than the editor limit."))
	}
	return packageFile{
		path:        filePath,
		kind:        kind,
		bytes:       content,
		sha256:      digest(content),
		contentType: contentTypeForPath(filePath),
		text:        text,
	}, issues, nil
}

// readZipEntry reads one zip entry with a hard cap.
func readZipEntry(ctx context.Context, entry *zip.File, limit int64) ([]byte, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	handle, err := entry.Open()
	if err != nil {
		return nil, fmt.Errorf("open zip entry %q: %w", entry.Name, err)
	}
	defer handle.Close()
	content, err := io.ReadAll(io.LimitReader(handle, limit+1))
	if err != nil {
		return nil, err
	}
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	return content, nil
}

// checkContext returns ctx cancellation when present.
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// ratioTooHigh reports suspicious compression ratios.
func ratioTooHigh(entry *zip.File, limit float64) bool {
	if entry.UncompressedSize64 == 0 || entry.CompressedSize64 == 0 {
		return entry.UncompressedSize64 > 0 && entry.CompressedSize64 == 0
	}
	return float64(entry.UncompressedSize64)/float64(entry.CompressedSize64) > limit
}

// digest returns a lowercase SHA-256 digest.
func digest(content []byte) domain.Digest {
	hash := sha256.Sum256(content)
	return domain.Digest(hex.EncodeToString(hash[:]))
}

// newIssue creates an import validation issue.
func newIssue(code domain.ValidationIssueCode, filePath domain.FilePath, message string) domain.ThemeValidationIssue {
	return domain.ThemeValidationIssue{
		ID:       uuid.New(),
		Severity: domain.SeverityError,
		Code:     code,
		Path:     filePath,
		Message:  message,
		Details:  []byte(`{}`),
	}
}

// decodeManifest returns manifest content and issues.
func decodeManifest(files []packageFile) (manifestPayload, []domain.ThemeValidationIssue) {
	raw := findFileBytes(files, "realmkit-theme.json")
	if len(raw) == 0 {
		return manifestPayload{raw: []byte(`{}`), files: files}, []domain.ThemeValidationIssue{
			newIssue(domain.IssueInvalidManifest, "realmkit-theme.json", "Theme package manifest is required."),
		}
	}
	var document manifestDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return manifestPayload{raw: raw, files: files}, []domain.ThemeValidationIssue{
			newIssue(domain.IssueInvalidManifest, "realmkit-theme.json", "Theme package manifest JSON is invalid."),
		}
	}
	return manifestPayload{raw: raw, document: document, files: files}, nil
}

// findFileBytes returns file bytes by normalized path.
func findFileBytes(files []packageFile, filePath domain.FilePath) []byte {
	normalized := domain.NormalizeFilePath(filePath)
	for _, file := range files {
		if file.path == normalized {
			return file.bytes
		}
	}
	return nil
}

// integrityFiles maps package files into version integrity inputs.
func integrityFiles(files []packageFile) []domain.IntegrityFile {
	values := make([]domain.IntegrityFile, 0, len(files))
	for _, file := range files {
		values = append(values, domain.IntegrityFile{Path: file.path, ContentSHA256: file.sha256})
	}
	return values
}

// hasError reports whether validation issues block draft status.
func hasError(issues []domain.ThemeValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == domain.SeverityError {
			return true
		}
	}
	return false
}

// coalesce returns the first non-empty value.
func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
