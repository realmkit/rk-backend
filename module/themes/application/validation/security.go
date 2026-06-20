package validation

import (
	"context"
	"path"
	"strings"

	cssadapter "github.com/realmkit/rk-backend/module/themes/adapter/css"
	jsadapter "github.com/realmkit/rk-backend/module/themes/adapter/javascript"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// validateCSSFiles validates CSS assets and embedded stylesheet blocks.
func validateCSSFiles(ctx context.Context, files []domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, file := range files {
		if err := checkContext(ctx); err != nil {
			return issues
		}
		if strings.EqualFold(path.Ext(string(file.Path)), ".css") {
			issues = append(issues, cssIssues(file.Path, cssadapter.Validate(file.ContentText))...)
		}
		for _, block := range extractBlock(file.ContentText, "rk_stylesheet", "endrk_stylesheet") {
			issues = append(issues, cssIssues(file.Path, cssadapter.Validate(block))...)
		}
	}
	return issues
}

// validateJavaScriptFiles validates JS assets and embedded JavaScript blocks.
func validateJavaScriptFiles(ctx context.Context, files []domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, file := range files {
		if err := checkContext(ctx); err != nil {
			return issues
		}
		if strings.EqualFold(path.Ext(string(file.Path)), ".js") {
			issues = append(issues, jsIssues(file.Path, jsadapter.Validate(file.ContentText))...)
		}
		for _, block := range extractBlock(file.ContentText, "rk_javascript", "endrk_javascript") {
			issues = append(issues, jsIssues(file.Path, jsadapter.Validate(block))...)
		}
	}
	return issues
}

// cssIssues maps CSS adapter issues to validation issues.
func cssIssues(filePath domain.FilePath, adapterIssues []cssadapter.Issue) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0, len(adapterIssues))
	for _, adapterIssue := range adapterIssues {
		code := domain.IssueInvalidCSS
		if adapterIssue.Code == "unsafe" {
			code = domain.IssueUnsafeCSS
		}
		issues = append(issues, issue(code, filePath, adapterIssue.Message))
	}
	return issues
}

// jsIssues maps JavaScript adapter issues to validation issues.
func jsIssues(filePath domain.FilePath, adapterIssues []jsadapter.Issue) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0, len(adapterIssues))
	for _, adapterIssue := range adapterIssues {
		code := domain.IssueInvalidJavaScript
		if adapterIssue.Code == "unsafe" {
			code = domain.IssueUnsafeJavaScript
		}
		issues = append(issues, issue(code, filePath, adapterIssue.Message))
	}
	return issues
}

// extractBlock extracts simple RealmKit Liquid block contents.
func extractBlock(content string, open string, close string) []string {
	blocks := make([]string, 0)
	startToken := "{% " + open + " %}"
	endToken := "{% " + close + " %}"
	for {
		start := strings.Index(content, startToken)
		if start < 0 {
			return blocks
		}
		rest := content[start+len(startToken):]
		end := strings.Index(rest, endToken)
		if end < 0 {
			return append(blocks, rest)
		}
		blocks = append(blocks, rest[:end])
		content = rest[end+len(endToken):]
	}
}
