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
