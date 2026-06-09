package domain

import (
	"mime"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// keyPattern matches lower snake keys used by namespaces.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// ValidateNamespace validates namespace.
func ValidateNamespace(field string, namespace Namespace) []Violation {
	value := strings.TrimSpace(string(namespace))
	if !keyPattern.MatchString(value) {
		return []Violation{{Field: field, Message: "must start with a lowercase letter and contain lowercase letters, numbers, or underscores"}}
	}
	return nil
}

// ValidatePath validates virtual folder path.
func ValidatePath(field string, path VirtualPath) []Violation {
	value := strings.TrimSpace(string(path))
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.Contains(value, "\\") {
		return []Violation{{Field: field, Message: "must be a relative slash-separated path"}}
	}
	segments := strings.Split(value, "/")
	if len(segments) > 8 {
		return []Violation{{Field: field, Message: "must contain at most 8 folders"}}
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." || len(segment) > 80 {
			return []Violation{{Field: field, Message: "contains an invalid folder segment"}}
		}
	}
	return nil
}

// ValidateFilename validates filename.
func ValidateFilename(field string, filename Filename) []Violation {
	value := strings.TrimSpace(string(filename))
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if len(value) > 160 || strings.ContainsAny(value, `/\`) || strings.Contains(value, "..") {
		return []Violation{{Field: field, Message: "must be a plain filename up to 160 characters"}}
	}
	for _, char := range value {
		if unicode.IsControl(char) {
			return []Violation{{Field: field, Message: "must not contain control characters"}}
		}
	}
	return nil
}

// ValidateVisibility validates visibility.
func ValidateVisibility(field string, visibility Visibility) []Violation {
	if slices.Contains([]Visibility{VisibilityPublic, VisibilityAuthenticated, VisibilityPrivate}, visibility) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateStatus validates status.
func ValidateStatus(field string, status Status) []Violation {
	if slices.Contains([]Status{StatusPendingUpload, StatusAvailable, StatusQuarantined, StatusFailed}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateContentType validates content type.
func ValidateContentType(field string, contentType string) []Violation {
	value := strings.TrimSpace(strings.ToLower(contentType))
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if _, _, err := mime.ParseMediaType(value); err != nil {
		return []Violation{{Field: field, Message: "must be a valid media type"}}
	}
	if slices.Contains([]string{"image/png", "image/jpeg", "image/webp", "image/gif", "application/pdf", "text/plain", "application/json"}, value) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// NormalizePath returns a canonical virtual path.
func NormalizePath(path VirtualPath) VirtualPath {
	value := strings.Trim(strings.TrimSpace(string(path)), "/")
	if value == "" {
		return ""
	}
	segments := strings.Split(value, "/")
	for index, segment := range segments {
		segments[index] = strings.TrimSpace(segment)
	}
	return VirtualPath(strings.Join(segments, "/"))
}

// NormalizeFilename returns a canonical filename.
func NormalizeFilename(filename Filename) Filename {
	return Filename(filepath.Base(strings.TrimSpace(string(filename))))
}
