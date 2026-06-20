package validation

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// compiledManifest builds the backend validation manifest.
func compiledManifest(
	version domain.ThemeVersion,
	files []domain.ThemeFile,
	issues []domain.ThemeValidationIssue,
) ([]byte, error) {
	report := map[string]any{
		"version_id":       version.ID,
		"route_coverage":   routeCoverage(files),
		"dependency_graph": dependencyGraph(files),
		"issue_count":      len(issues),
		"files":            manifestFiles(files),
	}
	return json.Marshal(report)
}

// routeCoverage returns all route template coverage entries.
func routeCoverage(files []domain.ThemeFile) []coverageEntry {
	index := indexFiles(files)
	entries := make([]coverageEntry, 0, len(domain.RouteKinds()))
	for _, route := range domain.RouteKinds() {
		filePath := routeTemplatePath(route)
		_, present := index[filePath]
		entries = append(entries, coverageEntry{Route: route, Path: filePath, Present: present})
	}
	return entries
}

// dependencyGraph returns a compact dependency report.
func dependencyGraph(files []domain.ThemeFile) dependencyReport {
	report := dependencyReport{}
	for _, file := range files {
		switch file.Kind {
		case domain.FileKindSection:
			report.Sections = append(report.Sections, keyFromPath(file.Path, "sections/"))
		case domain.FileKindSnippet:
			report.Snippets = append(report.Snippets, keyFromPath(file.Path, "snippets/"))
		case domain.FileKindAsset:
			report.Assets = append(report.Assets, string(file.Path))
		}
	}
	sort.Strings(report.Sections)
	sort.Strings(report.Snippets)
	sort.Strings(report.Assets)
	return report
}

// manifestFiles returns stable file summaries.
func manifestFiles(files []domain.ThemeFile) []map[string]any {
	values := make([]map[string]any, 0, len(files))
	for _, file := range files {
		values = append(values, map[string]any{
			"path":   file.Path,
			"kind":   file.Kind,
			"sha256": file.ContentSHA256,
			"size":   file.SizeBytes,
		})
	}
	sort.SliceStable(values, func(left int, right int) bool {
		return values[left]["path"].(domain.FilePath) < values[right]["path"].(domain.FilePath)
	})
	return values
}

// keyFromPath returns a logical key without extension.
func keyFromPath(filePath domain.FilePath, prefix string) string {
	value := string(domain.NormalizeFilePath(filePath))
	value = strings.TrimPrefix(value, prefix)
	return strings.TrimSuffix(value, ".liquid")
}
