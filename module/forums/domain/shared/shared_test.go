package shared

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// TestValidationHelpersAcceptValidValues covers happy-path validation helpers.
func TestValidationHelpersAcceptValidValues(t *testing.T) {
	validations := [][]Violation{
		ValidateKey("key", "forum_news"),
		ValidateSlug("slug", "forum-news"),
		ValidateName("name", "Forum News"),
		ValidateDescription("description", "Useful updates"),
		ValidateDisplayOrder("display_order", 0),
		ValidateCategoryStatus("status", CategoryStatusActive),
		ValidateForumKind("kind", ForumKindDiscussion),
		ValidateForumStatus("status", ForumStatusActive),
		ValidateThreadVisibilityMode("mode", ThreadVisibilityOwnOrStickyThreads),
		ValidateThreadStatus("status", ThreadStatusOpen),
		ValidateStickyState("sticky_state", StickyStateSticky),
		ValidatePostStatus("status", PostStatusVisible),
		ValidateContentFormat("content_format", ContentFormatProseMirror),
		ValidateReferenceType("reference_type", ReferenceQuote),
		ValidatePermissionSubjectType("subject_type", PermissionSubjectGroup),
		ValidateExternalURL("external_url", "https://example.com/rules"),
		ValidateTitle("title", "A useful thread"),
		ValidateContentDocument("document", json.RawMessage(`{"type":"doc"}`)),
		ValidateContentText("text", "hello"),
	}

	for _, violations := range validations {
		if len(violations) != 0 {
			t.Fatalf("expected valid value to pass, got %#v", violations)
		}
	}
}

// TestValidationHelpersRejectInvalidValues covers invalid validation branches.
func TestValidationHelpersRejectInvalidValues(t *testing.T) {
	invalidations := [][]Violation{
		ValidateKey("key", "Bad Key"),
		ValidateSlug("slug", "Bad Slug"),
		ValidateName("name", ""),
		ValidateDescription("description", strings.Repeat("x", 1001)),
		ValidateDisplayOrder("display_order", -1),
		ValidateCategoryStatus("status", "paused"),
		ValidateForumKind("kind", "chat"),
		ValidateForumStatus("status", "paused"),
		ValidateThreadVisibilityMode("mode", "secret"),
		ValidateThreadStatus("status", "paused"),
		ValidateStickyState("sticky_state", "floating"),
		ValidatePostStatus("status", "gone"),
		ValidateContentFormat("content_format", "html"),
		ValidateReferenceType("reference_type", "embed"),
		ValidatePermissionSubjectType("subject_type", "team"),
		ValidateExternalURL("external_url", ""),
		ValidateExternalURL("external_url", "/relative"),
		ValidateTitle("title", "no"),
		ValidateContentDocument("document", nil),
		ValidateContentDocument("document", json.RawMessage(`nope`)),
		ValidateContentDocument("document", json.RawMessage(`[]`)),
		ValidateContentDocument("document", json.RawMessage(`{`)),
		ValidateContentText("text", ""),
	}

	for _, violations := range invalidations {
		if len(violations) == 0 {
			t.Fatalf("expected invalid value to fail")
		}
	}
}

// TestValidationErrorAndRootObjectID covers shared error helpers and reserved ids.
func TestValidationErrorAndRootObjectID(t *testing.T) {
	err := NewValidationError(AppendViolation(nil, "field", "message"))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected validation error to wrap ErrInvalid")
	}
	if err := NewValidationError(nil); err != nil {
		t.Fatalf("expected nil error with no violations: %v", err)
	}
	if RootForumObjectID().String() != "00000000-0000-0000-0000-000000000101" {
		t.Fatalf("unexpected root forum object id")
	}
}
