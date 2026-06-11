package content

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain/shared"
)

// TestThreadNormalizeValidateAndStateRules covers thread defaults and reader rules.
func TestThreadNormalizeValidateAndStateRules(t *testing.T) {
	thread := Thread{
		ID:           uuid.New(),
		ForumID:      uuid.New(),
		AuthorUserID: uuid.New(),
		Title:        "  Welcome thread  ",
		Slug:         " welcome-thread ",
	}.Normalize()

	if thread.Status != ThreadStatusOpen {
		t.Fatalf("expected default open status, got %q", thread.Status)
	}
	if thread.StickyState != StickyStateNormal {
		t.Fatalf("expected default normal sticky state, got %q", thread.StickyState)
	}
	if thread.Version != 1 {
		t.Fatalf("expected default version 1, got %d", thread.Version)
	}
	if err := thread.Validate(); err != nil {
		t.Fatalf("expected normalized thread to validate: %v", err)
	}
	if !thread.Visible() || !thread.Replyable() {
		t.Fatalf("expected open thread to be visible and replyable")
	}

	thread.Status = ThreadStatusLocked
	if !thread.Visible() || thread.Replyable() {
		t.Fatalf("expected locked thread to be visible but not replyable")
	}

	thread.Status = ThreadStatus("hidden")
	if thread.Visible() || thread.Replyable() {
		t.Fatalf("expected hidden thread to be invisible and not replyable")
	}
}

// TestThreadValidateRejectsInvalidData covers invalid thread invariants.
func TestThreadValidateRejectsInvalidData(t *testing.T) {
	thread := Thread{
		Title:             "no",
		Slug:              "Bad Slug",
		Status:            "paused",
		StickyState:       "floating",
		StickyOrder:       -1,
		LockedReason:      strings.Repeat("x", 501),
		ReplyCount:        -1,
		VisibleReplyCount: -1,
		PostCount:         -1,
		VisiblePostCount:  -1,
		LikeCount:         -1,
		ViewCount:         -1,
	}

	if err := thread.Validate(); err == nil {
		t.Fatalf("expected invalid thread to fail validation")
	}
}

// TestPostNormalizeValidateAndVisibility covers post defaults and reader visibility.
func TestPostNormalizeValidateAndVisibility(t *testing.T) {
	post := Post{
		ID:                  uuid.New(),
		ThreadID:            uuid.New(),
		ForumID:             uuid.New(),
		AuthorUserID:        uuid.New(),
		Sequence:            1,
		ContentDocumentJSON: json.RawMessage(`{"type":"doc","content":[]}`),
		ContentText:         "  hello forum  ",
	}.Normalize()

	if post.Status != PostStatusVisible {
		t.Fatalf("expected default visible status, got %q", post.Status)
	}
	if post.ContentFormat != ContentFormatProseMirror {
		t.Fatalf("expected default ProseMirror format, got %q", post.ContentFormat)
	}
	if post.Version != 1 {
		t.Fatalf("expected default version 1, got %d", post.Version)
	}
	if post.ContentText != "hello forum" {
		t.Fatalf("expected trimmed text, got %q", post.ContentText)
	}
	if err := post.Validate(); err != nil {
		t.Fatalf("expected normalized post to validate: %v", err)
	}
	if !post.Visible() {
		t.Fatalf("expected visible post to be visible")
	}

	post.Status = PostStatusSystem
	if !post.Visible() {
		t.Fatalf("expected system post to be visible")
	}

	post.Status = PostStatus("hidden")
	if post.Visible() {
		t.Fatalf("expected hidden post to be invisible")
	}
}

// TestPostValidateRejectsInvalidData covers invalid post invariants.
func TestPostValidateRejectsInvalidData(t *testing.T) {
	post := Post{
		Sequence:            0,
		Status:              "gone",
		ContentFormat:       "html",
		ContentDocumentJSON: json.RawMessage(`[]`),
		ContentText:         "",
		ContentChecksum:     strings.Repeat("x", 129),
		EditCount:           -1,
		LikeCount:           -1,
		ReplyReferenceCount: -1,
	}

	if err := post.Validate(); err == nil {
		t.Fatalf("expected invalid post to fail validation")
	}
}

// TestPostReferenceValidateRequiresTypeSpecificTargets covers reference target rules.
func TestPostReferenceValidateRequiresTypeSpecificTargets(t *testing.T) {
	sourceID := uuid.New()
	targetPostID := uuid.New()
	targetUserID := uuid.New()
	targetAssetID := uuid.New()

	cases := []PostReference{
		{
			SourcePostID:  sourceID,
			TargetPostID:  &targetPostID,
			ReferenceType: ReferenceReplyTo,
		},
		{
			SourcePostID:  sourceID,
			TargetPostID:  &targetPostID,
			ReferenceType: ReferenceQuote,
			QuoteExcerpt:  "stable quote",
		},
		{
			SourcePostID:  sourceID,
			TargetUserID:  &targetUserID,
			ReferenceType: ReferenceMention,
		},
		{
			SourcePostID:  sourceID,
			TargetAssetID: &targetAssetID,
			ReferenceType: ReferenceAttachment,
		},
		{
			SourcePostID:  sourceID,
			ReferenceType: ReferenceLink,
			LinkURL:       "https://example.com/rules",
		},
	}

	for _, reference := range cases {
		if err := reference.Validate(); err != nil {
			t.Fatalf("expected valid %s reference: %v", reference.ReferenceType, err)
		}
	}

	invalidCases := []PostReference{
		{SourcePostID: sourceID, ReferenceType: ReferenceReplyTo},
		{SourcePostID: sourceID, ReferenceType: ReferenceQuote},
		{SourcePostID: sourceID, ReferenceType: ReferenceMention},
		{SourcePostID: sourceID, ReferenceType: ReferenceAttachment},
		{SourcePostID: sourceID, ReferenceType: ReferenceLink, LinkURL: "local-path"},
		{ReferenceType: ReferenceLink, LinkURL: "https://example.com"},
		{
			SourcePostID:  sourceID,
			TargetPostID:  &targetPostID,
			ReferenceType: ReferenceQuote,
			QuoteExcerpt:  strings.Repeat("x", 501),
		},
	}

	for _, reference := range invalidCases {
		if err := reference.Validate(); err == nil {
			t.Fatalf("expected invalid %s reference to fail", reference.ReferenceType)
		}
	}
}

// TestPostLikeAndReadStateValidation covers interaction identity validation.
func TestPostLikeAndReadStateValidation(t *testing.T) {
	like := PostLike{
		ID:       uuid.New(),
		PostID:   uuid.New(),
		ThreadID: uuid.New(),
		ForumID:  uuid.New(),
		UserID:   uuid.New(),
	}
	if err := like.Validate(); err != nil {
		t.Fatalf("expected like to validate: %v", err)
	}
	if err := (PostLike{}).Validate(); err == nil {
		t.Fatalf("expected empty like to fail validation")
	}

	state := ThreadReadState{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		ForumID:              uuid.New(),
		ThreadID:             uuid.New(),
		LastReadPostSequence: 2,
		LastReadAt:           time.Now(),
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("expected read state to validate: %v", err)
	}
	if err := (ThreadReadState{}).Validate(); err == nil {
		t.Fatalf("expected empty read state to fail validation")
	}
}

// TestAliasesReturnValidationErrors covers alias helpers used by wrapper packages.
func TestAliasesReturnValidationErrors(t *testing.T) {
	err := NewValidationError(AppendViolation(nil, "field", "message"))
	if !errors.Is(err, shared.ErrInvalid) {
		t.Fatalf("expected validation error to wrap ErrInvalid")
	}
	if err := NewValidationError(nil); err != nil {
		t.Fatalf("expected nil error with no violations: %v", err)
	}
}
