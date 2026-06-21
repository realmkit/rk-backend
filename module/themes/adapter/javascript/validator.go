// Package javascript validates theme JavaScript before publication.
package javascript

import "strings"

// Issue describes one JavaScript validation issue.
type Issue struct {
	Code    string // Code stores the code value.
	Message string // Message stores the message value.
}

// Validate checks first-version JavaScript safety rules.
func Validate(content string) []Issue {
	issues := make([]Issue, 0)
	lower := strings.ToLower(content)
	if strings.Contains(lower, "eval(") ||
		strings.Contains(lower, "new function") ||
		strings.Contains(lower, "import(") ||
		strings.Contains(lower, " from \"http") ||
		strings.Contains(lower, " from 'http") {
		issues = append(issues, Issue{Code: "unsafe", Message: "JavaScript must not evaluate strings or import remote code."})
	}
	if !balanced(content, '{', '}') || !balanced(content, '(', ')') {
		issues = append(issues, Issue{Code: "invalid", Message: "JavaScript braces or parentheses are not balanced."})
	}
	return issues
}

// balanced reports whether delimiters are balanced.
func balanced(value string, open rune, close rune) bool {
	depth := 0
	for _, char := range value {
		if char == open {
			depth++
		}
		if char == close {
			depth--
		}
		if depth < 0 {
			return false
		}
	}
	return depth == 0
}
