package validation

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

var (
	// realmKitTagPattern stores package state.
	realmKitTagPattern = regexp.MustCompile(`{%\s*(rk_[a-zA-Z0-9_]+)`)
	// referencePatterns stores package state.
	referencePatterns = []referencePattern{
		{pattern: regexp.MustCompile(`rk_section\s+['"]([^'"]+)['"]`), format: "sections/%s.liquid"},
		{pattern: regexp.MustCompile(`render\s+['"]([^'"]+)['"]`), format: "snippets/%s.liquid"},
		{pattern: regexp.MustCompile(`['"]([^'"]+)['"]\s*\|\s*asset_url`), format: "assets/%s"},
	}
)

// referencePattern describes one Liquid dependency matcher.
type referencePattern struct {
	pattern *regexp.Regexp // pattern stores the pattern value.
	format  string         // format stores the format value.
}

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
	ctx context.Context,
	files []domain.ThemeFile,
	index map[domain.FilePath]domain.ThemeFile,
) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, file := range files {
		if err := checkContext(ctx); err != nil {
			return issues
		}
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
func validateRouteCoverage(ctx context.Context, index map[domain.FilePath]domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, route := range domain.RouteKinds() {
		if err := checkContext(ctx); err != nil {
			return issues
		}
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
	matches := realmKitTagPattern.FindAllStringSubmatch(file.ContentText, -1)
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, match := range matches {
		if !allowedRealmKitTag(match[1]) {
			issues = append(issues, issue(domain.IssueUnknownRealmKitTag, file.Path, "Unknown RealmKit Liquid tag."))
		}
	}
	return issues
}

// allowedRealmKitTag reports whether a custom RealmKit tag is supported.
func allowedRealmKitTag(tag string) bool {
	switch tag {
	case "rk_layout", "rk_section", "rk_schema", "rk_stylesheet", "rk_javascript", "rk_doc":
		return true
	default:
		return false
	}
}

// dependencyIssues returns missing section, snippet, and asset diagnostics.
func dependencyIssues(file domain.ThemeFile, index map[domain.FilePath]domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, reference := range referencePatterns {
		issues = append(issues, missingReferences(file, index, reference)...)
	}
	return issues
}

// missingReferences returns missing dependency diagnostics.
func missingReferences(
	file domain.ThemeFile,
	index map[domain.FilePath]domain.ThemeFile,
	reference referencePattern,
) []domain.ThemeValidationIssue {
	matches := reference.pattern.FindAllStringSubmatch(file.ContentText, -1)
	issues := make([]domain.ThemeValidationIssue, 0, len(matches))
	for _, match := range matches {
		if _, ok := index[referencePath(reference.format, match[1])]; !ok {
			issues = append(issues, issue(domain.IssueMissingDependency, file.Path, "Theme file references a missing dependency."))
		}
	}
	return issues
}

// referencePath returns a normalized dependency path.
func referencePath(format string, value string) domain.FilePath {
	return domain.NormalizeFilePath(domain.FilePath(strings.ReplaceAll(format, "%s", value)))
}

// routeTemplatePath returns the required template path for a route.
func routeTemplatePath(route domain.RouteKind) domain.FilePath {
	return domain.FilePath("templates/" + string(route) + ".liquid")
}
