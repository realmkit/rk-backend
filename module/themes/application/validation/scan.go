package validation

import (
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// indexFiles indexes files by path and kind-derived key.
func indexFiles(files []domain.ThemeFile) map[domain.FilePath]domain.ThemeFile {
	index := map[domain.FilePath]domain.ThemeFile{}
	for _, file := range files {
		index[domain.NormalizeFilePath(file.Path)] = file
	}
	return index
}

// validateRequiredStructure verifies first-version required folders.
func validateRequiredStructure(index map[domain.FilePath]domain.ThemeFile) []domain.ThemeValidationIssue {
	if hasPrefix(index, "layout/") && hasPrefix(index, "templates/") {
		return nil
	}
	issues := make([]domain.ThemeValidationIssue, 0)
	if !hasPrefix(index, "layout/") {
		issues = append(issues, issue(domain.IssueMissingRequiredDirectory, "layout", "Theme package requires a layout directory."))
	}
	if !hasPrefix(index, "templates/") {
		issues = append(issues, issue(domain.IssueMissingRequiredDirectory, "templates", "Theme package requires a templates directory."))
	}
	return issues
}

// validateLiquidFiles validates Liquid syntax and dependencies.
func validateLiquidFiles(
	files []domain.ThemeFile,
	index map[domain.FilePath]domain.ThemeFile,
) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, file := range files {
		if !isLiquid(file) {
			continue
		}
		content := file.ContentText
		if strings.Count(content, "{{") != strings.Count(content, "}}") ||
			strings.Count(content, "{%") != strings.Count(content, "%}") {
			issues = append(issues, issue(domain.IssueInvalidLiquid, file.Path, "Liquid delimiters are not balanced."))
		}
		issues = append(issues, unknownRKTagIssues(file)...)
		issues = append(issues, dependencyIssues(file, index)...)
	}
	return issues
}

// validateRouteCoverage verifies every route has a route template.
func validateRouteCoverage(index map[domain.FilePath]domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, route := range domain.RouteKinds() {
		filePath := routeTemplatePath(route)
		if _, ok := index[filePath]; !ok {
			issues = append(issues, issue(domain.IssueMissingRequiredTemplate, filePath, "Required route template is missing."))
		}
	}
	return issues
}

// issue creates one static validation issue.
func issue(code domain.ValidationIssueCode, filePath domain.FilePath, message string) domain.ThemeValidationIssue {
	return domain.ThemeValidationIssue{
		ID:       uuid.New(),
		Severity: domain.SeverityError,
		Code:     code,
		Path:     domain.NormalizeFilePath(filePath),
		Message:  message,
		Details:  []byte(`{}`),
	}
}

// hasPrefix reports whether indexed files include one prefix.
func hasPrefix(index map[domain.FilePath]domain.ThemeFile, prefix string) bool {
	for filePath := range index {
		if strings.HasPrefix(string(filePath), prefix) {
			return true
		}
	}
	return false
}

// isLiquid reports whether a file is a Liquid source.
func isLiquid(file domain.ThemeFile) bool {
	return slices.Contains(
		[]domain.FileKind{domain.FileKindLayout, domain.FileKindTemplate, domain.FileKindSection, domain.FileKindSnippet},
		file.Kind,
	)
}

// unknownRKTagIssues returns unsupported RealmKit tag diagnostics.
func unknownRKTagIssues(file domain.ThemeFile) []domain.ThemeValidationIssue {
	matches := regexp.MustCompile(`{%\s*(rk_[a-zA-Z0-9_]+)`).FindAllStringSubmatch(file.ContentText, -1)
	allowed := map[string]struct{}{"rk_layout": {}, "rk_section": {}, "rk_schema": {}, "rk_stylesheet": {}, "rk_javascript": {}, "rk_doc": {}}
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, match := range matches {
		if _, ok := allowed[match[1]]; !ok {
			issues = append(issues, issue(domain.IssueUnknownRealmKitTag, file.Path, "Unknown RealmKit Liquid tag."))
		}
	}
	return issues
}

// dependencyIssues returns missing section, snippet, and asset diagnostics.
func dependencyIssues(file domain.ThemeFile, index map[domain.FilePath]domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	issues = append(issues, missingReferences(file, index, `rk_section\s+['"]([^'"]+)['"]`, "sections/%s.liquid")...)
	issues = append(issues, missingReferences(file, index, `render\s+['"]([^'"]+)['"]`, "snippets/%s.liquid")...)
	issues = append(issues, missingReferences(file, index, `['"]([^'"]+)['"]\s*\|\s*asset_url`, "assets/%s")...)
	return issues
}

// missingReferences returns missing dependency diagnostics.
func missingReferences(
	file domain.ThemeFile,
	index map[domain.FilePath]domain.ThemeFile,
	pattern string,
	format string,
) []domain.ThemeValidationIssue {
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(file.ContentText, -1)
	issues := make([]domain.ThemeValidationIssue, 0, len(matches))
	for _, match := range matches {
		if _, ok := index[domain.FilePath(strings.ReplaceAll(format, "%s", match[1]))]; !ok {
			issues = append(issues, issue(domain.IssueMissingDependency, file.Path, "Theme file references a missing dependency."))
		}
	}
	return issues
}

// routeTemplatePath returns the required template path for a route.
func routeTemplatePath(route domain.RouteKind) domain.FilePath {
	return domain.FilePath("templates/" + string(route) + ".liquid")
}
