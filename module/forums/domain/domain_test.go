package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestCategoryValidateAcceptsDefaults verifies category normalization.
func TestCategoryValidateAcceptsDefaults(t *testing.T) {
	category := ForumCategory{Key: "cube_official", Name: "CubeCraft Official"}.Normalize()
	if err := category.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if category.Status != CategoryStatusActive || category.Version != 1 {
		t.Fatalf("category = %+v, want active version 1", category)
	}
}

// TestForumValidateRejectsInvalidLink verifies link forum URL validation.
func TestForumValidateRejectsInvalidLink(t *testing.T) {
	forum := Forum{ID: uuid.New(), CategoryID: uuid.New(), Kind: ForumKindLink, Key: "discord", Slug: "discord", Name: "Discord", Path: "/" + uuid.NewString() + "/", Status: ForumStatusActive}.Normalize()
	forum.Path = "/" + forum.ID.String() + "/"
	err := forum.Validate()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestForumValidateAcceptsDiscussion verifies discussion forum validation.
func TestForumValidateAcceptsDiscussion(t *testing.T) {
	id := uuid.New()
	forum := Forum{ID: id, CategoryID: uuid.New(), Key: "news", Slug: "news", Name: "News", Path: "/" + id.String() + "/"}.Normalize()
	if err := forum.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !forum.Discussion() {
		t.Fatalf("Discussion() = false, want true")
	}
}

// TestRootForumObjectIDIsStable verifies reserved permission target stability.
func TestRootForumObjectIDIsStable(t *testing.T) {
	if RootForumObjectID().String() != "00000000-0000-0000-0000-000000000101" {
		t.Fatalf("RootForumObjectID() = %s", RootForumObjectID())
	}
}

// TestPostValidateRejectsInvalidContent verifies WYSIWYG JSON validation.
func TestPostValidateRejectsInvalidContent(t *testing.T) {
	post := Post{ID: uuid.New(), ThreadID: uuid.New(), ForumID: uuid.New(), AuthorUserID: uuid.New(), Sequence: 1, ContentDocumentJSON: []byte(`[]`), ContentText: "hello"}.Normalize()
	err := post.Validate()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestThreadAndPostValidateAcceptValidContent verifies content entities accept valid state.
func TestThreadAndPostValidateAcceptValidContent(t *testing.T) {
	authorID := uuid.New()
	thread := Thread{ID: uuid.New(), ForumID: uuid.New(), AuthorUserID: authorID, OpenerPostID: uuid.New(), LatestPostID: uuid.New(), LatestPostAuthorUserID: authorID, Title: "Valid title", Slug: "valid-title"}.Normalize()
	if err := thread.Validate(); err != nil {
		t.Fatalf("Thread Validate() error = %v", err)
	}
	post := Post{ID: uuid.New(), ThreadID: thread.ID, ForumID: thread.ForumID, AuthorUserID: authorID, Sequence: 1, ContentDocumentJSON: []byte(`{"type":"doc"}`), ContentText: "hello"}.Normalize()
	if err := post.Validate(); err != nil {
		t.Fatalf("Post Validate() error = %v", err)
	}
	if !thread.Visible() || !thread.Replyable() || !post.Visible() {
		t.Fatalf("thread/post visibility helpers returned false")
	}
}

// TestPostReferenceValidateRequiresTargets verifies structured reference validation.
func TestPostReferenceValidateRequiresTargets(t *testing.T) {
	reference := PostReference{ID: uuid.New(), SourcePostID: uuid.New(), ReferenceType: ReferenceAttachment}
	err := reference.Validate()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
	assetID := uuid.New()
	reference.TargetAssetID = &assetID
	if err := reference.Validate(); err != nil {
		t.Fatalf("Validate() valid attachment error = %v", err)
	}
}
