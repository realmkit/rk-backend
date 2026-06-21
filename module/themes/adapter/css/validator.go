// Package css validates theme CSS before publication.
package css

import (
	"strings"
)

// Issue describes one CSS validation issue.
type Issue struct {
	Code    string // Code stores the code value.
	Message string // Message stores the message value.
}

// Validate checks first-version CSS safety rules.
func Validate(content string) []Issue {
	issues := make([]Issue, 0)
	lower := strings.ToLower(content)
	if strings.Contains(lower, "@import") || strings.Contains(lower, "url(http") || strings.Contains(lower, "url(//") {
		issues = append(issues, Issue{Code: "unsafe", Message: "CSS must not import remote resources."})
	}
	if !balanced(content, '{', '}') || !balanced(content, '(', ')') {
		issues = append(issues, Issue{Code: "invalid", Message: "CSS braces or parentheses are not balanced."})
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
