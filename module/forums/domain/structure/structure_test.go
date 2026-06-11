package structure

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestForumCategoryNormalizeAndValidate covers category defaults and validation.
func TestForumCategoryNormalizeAndValidate(t *testing.T) {
	category := ForumCategory{
		Key:         " official ",
		Name:        "  Official ",
		Description: "  Announcements ",
	}.Normalize()

	if category.Status != CategoryStatusActive {
		t.Fatalf("expected default active status, got %q", category.Status)
	}
	if category.Version != 1 {
		t.Fatalf("expected default version 1, got %d", category.Version)
	}
	if category.Key != "official" || category.Name != "Official" {
		t.Fatalf("expected trimmed category fields: %#v", category)
	}
	if err := category.Validate(); err != nil {
		t.Fatalf("expected normalized category to validate: %v", err)
	}

	invalid := ForumCategory{
		Key:          "Bad Key",
		Name:         "",
		Description:  strings.Repeat("x", 1001),
		DisplayOrder: -1,
		Status:       "paused",
	}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected invalid category to fail validation")
	}
}

// TestForumNormalizeValidateAndSettings covers forum defaults and settings mapping.
func TestForumNormalizeValidateAndSettings(t *testing.T) {
	forumID := uuid.New()
	categoryID := uuid.New()
	forum := Forum{
		ID:          forumID,
		CategoryID:  categoryID,
		Key:         " news ",
		Slug:        " news-board ",
		Name:        " News ",
		Description: " Updates ",
		Path:        "/" + forumID.String() + "/",
	}.Normalize()

	if forum.Kind != ForumKindDiscussion {
		t.Fatalf("expected default discussion kind, got %q", forum.Kind)
	}
	if forum.ThreadVisibilityMode != ThreadVisibilityAllThreads {
		t.Fatalf("expected default visibility mode, got %q", forum.ThreadVisibilityMode)
	}
	if forum.DefaultThreadStatus != ThreadStatusOpen {
		t.Fatalf("expected default thread status, got %q", forum.DefaultThreadStatus)
	}
	if forum.Status != ForumStatusActive {
		t.Fatalf("expected default active status, got %q", forum.Status)
	}
	if forum.Version != 1 {
		t.Fatalf("expected default version 1, got %d", forum.Version)
	}
	if !forum.Discussion() {
		t.Fatalf("expected discussion forum to accept threads")
	}
	if err := forum.Validate(); err != nil {
		t.Fatalf("expected normalized forum to validate: %v", err)
	}

	settings := forum.Settings()
	if settings.ForumID != forum.ID || settings.Version != forum.Version {
		t.Fatalf("expected settings to mirror forum identity/version: %#v", settings)
	}
}

// TestForumValidateRejectsInvalidBehavior covers path, kind, and window rules.
func TestForumValidateRejectsInvalidBehavior(t *testing.T) {
	forumID := uuid.New()
	invalid := Forum{
		ID:                            forumID,
		Key:                           "bad key",
		Slug:                          "Bad Slug",
		Name:                          "",
		Description:                   strings.Repeat("x", 1001),
		DisplayOrder:                  -1,
		Kind:                          "chat",
		Status:                        "paused",
		ThreadVisibilityMode:          "secret",
		DefaultThreadStatus:           "paused",
		MaxStickyThreads:              -1,
		AuthorPostEditWindowSeconds:   -2,
		AuthorPostDeleteWindowSeconds: -2,
		Path:                          "/" + uuid.NewString() + "/",
		Depth:                         6,
	}

	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected invalid forum to fail validation")
	}

	link := Forum{
		ID:          forumID,
		CategoryID:  uuid.New(),
		Kind:        ForumKindLink,
		Key:         "discord",
		Slug:        "discord",
		Name:        "Discord",
		Path:        "/" + forumID.String() + "/",
		ExternalURL: "https://discord.example.com",
	}.Normalize()
	if err := link.Validate(); err != nil {
		t.Fatalf("expected valid link forum: %v", err)
	}
	if link.Discussion() {
		t.Fatalf("expected link forum not to accept threads")
	}

	link.ExternalURL = ""
	if err := link.Validate(); err == nil {
		t.Fatalf("expected link forum without URL to fail")
	}

	discussionWithURL := link
	discussionWithURL.Kind = ForumKindDiscussion
	discussionWithURL.ExternalURL = "https://example.com"
	if err := discussionWithURL.Validate(); err == nil {
		t.Fatalf("expected non-link forum with URL to fail")
	}
}

// TestForumSettingsNormalizeAndValidate covers admin-editable settings rules.
func TestForumSettingsNormalizeAndValidate(t *testing.T) {
	settings := ForumSettings{
		ForumID:     uuid.New(),
		ExternalURL: " https://example.com/help ",
		Kind:        ForumKindLink,
	}.Normalize()

	if settings.ThreadVisibilityMode != ThreadVisibilityAllThreads {
		t.Fatalf("expected default visibility mode, got %q", settings.ThreadVisibilityMode)
	}
	if settings.DefaultThreadStatus != ThreadStatusOpen {
		t.Fatalf("expected default thread status, got %q", settings.DefaultThreadStatus)
	}
	if settings.ExternalURL != "https://example.com/help" {
		t.Fatalf("expected trimmed external url, got %q", settings.ExternalURL)
	}
	if err := settings.Validate(); err != nil {
		t.Fatalf("expected valid settings: %v", err)
	}

	invalid := ForumSettings{
		Kind:                          ForumKindDiscussion,
		ExternalURL:                   "https://example.com",
		ThreadVisibilityMode:          "secret",
		DefaultThreadStatus:           "paused",
		MaxStickyThreads:              -1,
		AuthorPostEditWindowSeconds:   -2,
		AuthorPostDeleteWindowSeconds: -2,
	}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected invalid settings to fail")
	}
}
