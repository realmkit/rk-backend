package delivery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// buildManifest assembles a Next.js render manifest.
func buildManifest(
	theme domain.Theme,
	version domain.ThemeVersion,
	files []domain.ThemeFile,
	assets []domain.ThemeAsset,
) ManifestResult {
	result := ManifestResult{
		Theme:           theme,
		Version:         version,
		Files:           files,
		Assets:          assetMap(assets),
		Layouts:         filesByKind(files, domain.FileKindLayout),
		Templates:       filesByKind(files, domain.FileKindTemplate),
		Sections:        filesByKind(files, domain.FileKindSection),
		Snippets:        filesByKind(files, domain.FileKindSnippet),
		Locales:         filesByKind(files, domain.FileKindLocale),
		SettingsSchema:  jsonObject(version.SettingsSchemaJSON),
		SettingsData:    jsonObject(version.SettingsDataJSON),
		DependencyGraph: manifestObject(version.ManifestJSON, "dependency_graph"),
		RouteCoverage:   manifestObject(version.ManifestJSON, "route_coverage"),
		CSP:             cspManifest(assets),
		Cache:           versionCache(version),
	}
	return result
}

// filesByKind indexes files by logical key.
func filesByKind(files []domain.ThemeFile, kind domain.FileKind) map[string]domain.ThemeFile {
	values := map[string]domain.ThemeFile{}
	for _, file := range files {
		if file.Kind == kind {
			values[strings.TrimSuffix(string(file.Path), ".liquid")] = file
		}
	}
	return values
}

// assetMap returns immutable asset metadata.
func assetMap(assets []domain.ThemeAsset) map[string]map[string]any {
	values := map[string]map[string]any{}
	for _, asset := range assets {
		values[string(asset.Path)] = map[string]any{
			"url":       asset.PublicURL,
			"type":      asset.ContentType,
			"sha256":    asset.ContentSHA256,
			"integrity": asset.IntegrityValue,
			"bytes":     asset.SizeBytes,
		}
	}
	return values
}

// cspManifest returns strict CSP asset directives.
func cspManifest(assets []domain.ThemeAsset) map[string]any {
	scripts := make([]string, 0)
	styles := make([]string, 0)
	for _, asset := range assets {
		if strings.Contains(asset.ContentType, "javascript") {
			scripts = append(scripts, asset.IntegrityValue)
		}
		if strings.Contains(asset.ContentType, "css") {
			styles = append(styles, asset.IntegrityValue)
		}
	}
	return map[string]any{
		"default_src": []string{"'self'"},
		"script_src":  append([]string{"'self'"}, scripts...),
		"style_src":   append([]string{"'self'"}, styles...),
		"img_src":     []string{"'self'", "data:"},
	}
}

// jsonObject decodes object JSON.
func jsonObject(content []byte) map[string]any {
	var value map[string]any
	if err := json.Unmarshal(content, &value); err != nil || value == nil {
		return map[string]any{}
	}
	return value
}

// manifestObject extracts one object from compiled manifest JSON.
func manifestObject(content []byte, key string) map[string]any {
	value := jsonObject(content)
	nested, ok := value[key].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return nested
}

// digestETag returns an HTTP ETag from arbitrary parts.
func digestETag(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return `"` + hex.EncodeToString(hash.Sum(nil)) + `"`
}

// activeCache returns cache metadata for the active pointer.
func activeCache(activation domain.ThemeActivation) CacheMetadata {
	return CacheMetadata{
		ETag:         digestETag(activation.ID.String(), activation.VersionID.String(), string(activation.SettingsDataJSON)),
		CacheControl: revalidateCacheControl,
		LastModified: activation.ActivatedAt,
	}
}

// versionCache returns cache metadata for immutable version data.
func versionCache(version domain.ThemeVersion) CacheMetadata {
	cacheControl := immutableCacheControl
	if version.Status == domain.VersionStatusDraft || version.Status == domain.VersionStatusValidating {
		cacheControl = noStoreCacheControl
	}
	return CacheMetadata{
		ETag:          digestETag(version.ID.String(), string(version.IntegritySHA256), string(version.ManifestJSON)),
		CacheControl:  cacheControl,
		LastModified:  latest(version.UpdatedAt, version.CreatedAt),
		ContentSHA256: version.IntegritySHA256,
	}
}

// fileCache returns cache metadata for one file.
func fileCache(version domain.ThemeVersion, file domain.ThemeFile) CacheMetadata {
	cache := versionCache(version)
	cache.ETag = `"` + string(file.ContentSHA256) + `"`
	cache.LastModified = latest(file.UpdatedAt, file.CreatedAt)
	cache.ContentSHA256 = file.ContentSHA256
	return cache
}

// assetCache returns cache metadata for one immutable asset.
func assetCache(asset domain.ThemeAsset) CacheMetadata {
	return CacheMetadata{
		ETag:          `"` + string(asset.ContentSHA256) + `"`,
		CacheControl:  immutableCacheControl,
		LastModified:  latest(asset.UpdatedAt, asset.CreatedAt),
		ContentSHA256: asset.ContentSHA256,
	}
}

// latest returns the first non-zero time.
func latest(values ...time.Time) time.Time {
	for _, value := range values {
		if !value.IsZero() {
			return value
		}
	}
	return time.Time{}
}
