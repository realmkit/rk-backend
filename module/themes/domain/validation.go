package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"sort"
	"strings"
)

// IntegrityFile is one path and content digest pair used by version hashing.
type IntegrityFile struct {
	Path          FilePath
	ContentSHA256 Digest
}

const (
	// IssuePackageTooLarge reports an uploaded package above size limits.
	IssuePackageTooLarge ValidationIssueCode = "package.too_large"
	// IssueExtractedPackageTooLarge reports extracted package size above limits.
	IssueExtractedPackageTooLarge ValidationIssueCode = "package.extracted_too_large"
	// IssueFileCountTooLarge reports too many package files.
	IssueFileCountTooLarge ValidationIssueCode = "package.too_many_files"
	// IssueTextFileTooLarge reports an editor-backed file above limits.
	IssueTextFileTooLarge ValidationIssueCode = "file.text_too_large"
	// IssueUnsafePath reports absolute or traversing paths.
	IssueUnsafePath ValidationIssueCode = "file.unsafe_path"
	// IssueDuplicatePath reports duplicate normalized paths.
	IssueDuplicatePath ValidationIssueCode = "file.duplicate_path"
	// IssueInvalidManifest reports malformed package manifest JSON.
	IssueInvalidManifest ValidationIssueCode = "manifest.invalid"
	// IssueInvalidSignature reports malformed or tampered signature data.
	IssueInvalidSignature ValidationIssueCode = "signature.invalid"
	// IssueUntrustedSignature reports an unknown signing key.
	IssueUntrustedSignature ValidationIssueCode = "signature.untrusted"
	// IssueRevokedSignature reports a revoked signing key.
	IssueRevokedSignature ValidationIssueCode = "signature.revoked"
	// IssueMissingRequiredDirectory reports a required theme directory is absent.
	IssueMissingRequiredDirectory ValidationIssueCode = "structure.missing_directory"
	// IssueMissingRequiredTemplate reports a required route template is absent.
	IssueMissingRequiredTemplate ValidationIssueCode = "template.missing_required"
	// IssueInvalidLiquid reports Liquid syntax errors.
	IssueInvalidLiquid ValidationIssueCode = "liquid.invalid_syntax"
	// IssueUnknownRealmKitTag reports unsupported custom `rk_` tags.
	IssueUnknownRealmKitTag ValidationIssueCode = "liquid.unknown_rk_tag"
	// IssueMissingDependency reports missing layout, snippet, section, or asset references.
	IssueMissingDependency ValidationIssueCode = "dependency.missing"
	// IssueInvalidSettingsSchema reports invalid settings schema JSON.
	IssueInvalidSettingsSchema ValidationIssueCode = "settings.schema_invalid"
	// IssueInvalidSettingsData reports invalid settings data JSON.
	IssueInvalidSettingsData ValidationIssueCode = "settings.data_invalid"
	// IssueInvalidLocale reports invalid locale JSON.
	IssueInvalidLocale ValidationIssueCode = "locale.invalid"
	// IssueInvalidCSS reports CSS syntax errors.
	IssueInvalidCSS ValidationIssueCode = "css.invalid"
	// IssueUnsafeCSS reports unsafe CSS imports or URLs.
	IssueUnsafeCSS ValidationIssueCode = "css.unsafe"
	// IssueInvalidJavaScript reports JavaScript syntax errors.
	IssueInvalidJavaScript ValidationIssueCode = "javascript.invalid"
	// IssueUnsafeJavaScript reports unsafe JavaScript APIs or imports.
	IssueUnsafeJavaScript ValidationIssueCode = "javascript.unsafe"
	// IssueRouteDataUnavailable reports a route requirement not supplied by Go.
	IssueRouteDataUnavailable ValidationIssueCode = "route_data.unavailable"
)

// ValidationIssueCodes returns all first-version validation issue codes.
func ValidationIssueCodes() []ValidationIssueCode {
	return []ValidationIssueCode{
		IssuePackageTooLarge,
		IssueExtractedPackageTooLarge,
		IssueFileCountTooLarge,
		IssueTextFileTooLarge,
		IssueUnsafePath,
		IssueDuplicatePath,
		IssueInvalidManifest,
		IssueInvalidSignature,
		IssueUntrustedSignature,
		IssueRevokedSignature,
		IssueMissingRequiredDirectory,
		IssueMissingRequiredTemplate,
		IssueInvalidLiquid,
		IssueUnknownRealmKitTag,
		IssueMissingDependency,
		IssueInvalidSettingsSchema,
		IssueInvalidSettingsData,
		IssueInvalidLocale,
		IssueInvalidCSS,
		IssueUnsafeCSS,
		IssueInvalidJavaScript,
		IssueUnsafeJavaScript,
		IssueRouteDataUnavailable,
	}
}

// CalculateVersionIntegritySHA256 returns a stable digest for version contents.
func CalculateVersionIntegritySHA256(files []IntegrityFile) Digest {
	normalized := make([]IntegrityFile, 0, len(files))
	for _, file := range files {
		normalized = append(normalized, IntegrityFile{
			Path:          NormalizeFilePath(file.Path),
			ContentSHA256: Digest(strings.ToLower(strings.TrimSpace(string(file.ContentSHA256)))),
		})
	}
	sort.SliceStable(normalized, func(left int, right int) bool {
		if normalized[left].Path == normalized[right].Path {
			return normalized[left].ContentSHA256 < normalized[right].ContentSHA256
		}
		return normalized[left].Path < normalized[right].Path
	})
	hash := sha256.New()
	for _, file := range normalized {
		hash.Write([]byte(file.Path))
		hash.Write([]byte{0})
		hash.Write([]byte(file.ContentSHA256))
		hash.Write([]byte{'\n'})
	}
	return Digest(hex.EncodeToString(hash.Sum(nil)))
}

// NormalizeFilePath returns a slash-separated package path without root escapes.
func NormalizeFilePath(filePath FilePath) FilePath {
	cleaned := strings.TrimSpace(strings.ReplaceAll(string(filePath), "\\", "/"))
	cleaned = strings.TrimPrefix(path.Clean("/"+cleaned), "/")
	if cleaned == "." {
		return ""
	}
	return FilePath(cleaned)
}
