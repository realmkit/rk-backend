package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestAssetValidateAcceptsValidAsset verifies a complete asset passes validation.
func TestAssetValidateAcceptsValidAsset(t *testing.T) {
	asset := validAsset()

	if err := asset.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestAssetValidateRejectsInvalidFields verifies validation returns all relevant failures.
func TestAssetValidateRejectsInvalidFields(t *testing.T) {
	asset := Asset{
		Namespace:   "Bad",
		Path:        "/bad",
		Filename:    "../bad",
		Visibility:  "unknown",
		Status:      "unknown",
		ContentType: "invalid",
	}

	err := asset.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 7 {
		t.Fatalf("Violations = %d, want at least 7", len(validation.Violations))
	}
}

// TestNormalizePathAndFilename verifies paths and filenames are canonicalized.
func TestNormalizePathAndFilename(t *testing.T) {
	if got := NormalizePath(" /media/icons/ "); got != "media/icons" {
		t.Fatalf("NormalizePath() = %q, want media/icons", got)
	}
	if got := NormalizeFilename("/tmp/Icon.png"); got != "Icon.png" {
		t.Fatalf("NormalizeFilename() = %q, want Icon.png", got)
	}
}

// TestAssetNormalizeDefaultsDisplayName verifies normalize fills display name.
func TestAssetNormalizeDefaultsDisplayName(t *testing.T) {
	asset := validAsset()
	asset.DisplayName = ""
	asset.Filename = " logo.png "

	normalized := asset.Normalize()
	if normalized.DisplayName != "logo.png" {
		t.Fatalf("DisplayName = %q, want logo.png", normalized.DisplayName)
	}
}

// TestValidationErrorMatchesInvalid verifies validation errors match ErrInvalid.
func TestValidationErrorMatchesInvalid(t *testing.T) {
	err := ValidationError{Violations: []Violation{{Field: "field", Message: "bad"}}}
	if !errors.Is(err, ErrInvalid) || err.Error() != ErrInvalid.Error() {
		t.Fatalf("ValidationError does not match ErrInvalid")
	}
}

// TestValidateContentTypeAppliesAllowlist verifies supported media types are accepted.
func TestValidateContentTypeAppliesAllowlist(t *testing.T) {
	if violations := ValidateContentType("content_type", "application/octet-stream"); len(violations) != 1 {
		t.Fatalf("ValidateContentType() violations = %d, want 1", len(violations))
	}
	if violations := ValidateContentType("content_type", "application/pdf"); len(violations) != 0 {
		t.Fatalf("ValidateContentType() violations = %d, want 0", len(violations))
	}
}

// validAsset returns a valid domain asset.
func validAsset() Asset {
	return Asset{
		ID:          uuid.New(),
		Namespace:   "community",
		Path:        "brand/icons",
		Filename:    "logo.png",
		DisplayName: "Logo",
		Visibility:  VisibilityPublic,
		Status:      StatusPendingUpload,
		StorageKey:  "assets/community/2026/06/id/logo.png",
		Bucket:      "gamehub-assets",
		ContentType: "image/png",
		SizeBytes:   128,
		Version:     1,
	}
}
