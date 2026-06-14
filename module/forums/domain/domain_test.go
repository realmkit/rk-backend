package domain

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

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
	forum := Forum{
		ID:         uuid.New(),
		CategoryID: uuid.New(),
		Kind:       ForumKindLink,
		Key:        "discord",
		Slug:       "discord",
		Name:       "Discord",
		Path:       "/" + uuid.NewString() + "/",
		Status:     ForumStatusActive,
	}.Normalize()
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

// TestForumSettingsReflectForumConfiguration verifies settings projection.
func TestForumSettingsReflectForumConfiguration(t *testing.T) {
	id := uuid.New()
	forum := Forum{
		ID:                            id,
		CategoryID:                    uuid.New(),
		Key:                           "support",
		Slug:                          "support",
		Name:                          "Support",
		Path:                          "/" + id.String() + "/",
		ThreadVisibilityMode:          ThreadVisibilityOwnThreads,
		MaxStickyThreads:              4,
		AuthorPostEditWindowSeconds:   45,
		AuthorPostDeleteWindowSeconds: -1,
	}.Normalize()

	settings := forum.Settings()

	if settings.ForumID != forum.ID || settings.ThreadVisibilityMode != ThreadVisibilityOwnThreads || settings.MaxStickyThreads != 4 ||
		settings.AuthorPostDeleteWindowSeconds != -1 {
		t.Fatalf("Settings() = %+v, want forum configuration", settings)
	}
}

// TestForumSettingsValidateSupportsDisabledAuthorWindows verifies settings validation.
func TestForumSettingsValidateSupportsDisabledAuthorWindows(t *testing.T) {
	settings := ForumSettings{
		ForumID:                       uuid.New(),
		Kind:                          ForumKindDiscussion,
		ThreadVisibilityMode:          ThreadVisibilityOwnOrStickyThreads,
		DefaultThreadStatus:           ThreadStatusOpen,
		MaxStickyThreads:              5,
		AuthorPostEditWindowSeconds:   -1,
		AuthorPostDeleteWindowSeconds: -1,
	}.Normalize()
	if err := settings.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestForumSettingsValidateRejectsExternalURLOnDiscussion verifies link-only URLs.
func TestForumSettingsValidateRejectsExternalURLOnDiscussion(t *testing.T) {
	settings := ForumSettings{
		ForumID:              uuid.New(),
		Kind:                 ForumKindDiscussion,
		ExternalURL:          "https://example.test",
		ThreadVisibilityMode: ThreadVisibilityAllThreads,
		DefaultThreadStatus:  ThreadStatusOpen,
	}.Normalize()

	err := settings.Validate()

	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestForumPermissionSettingsNormalizeReservedSubjects verifies grant normalization.
func TestForumPermissionSettingsNormalizeReservedSubjects(t *testing.T) {
	settings := ForumPermissionSettings{
		ForumID:    uuid.New(),
		Viewers:    []ForumPermissionGrant{{SubjectType: PermissionSubjectPublic}},
		Replyers:   []ForumPermissionGrant{{SubjectType: PermissionSubjectAuthenticated}},
		Moderators: []ForumPermissionGrant{{SubjectType: PermissionSubjectGroup, SubjectID: uuid.New()}},
	}.Normalize()
	if err := settings.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if settings.Viewers[0].SubjectID != PublicPermissionSubjectID() ||
		settings.Replyers[0].SubjectID != AuthenticatedPermissionSubjectID() {
		t.Fatalf("settings = %+v, want reserved ids", settings)
	}
}

// TestForumPermissionSettingsRejectsInvalidGrants verifies subject validation.
func TestForumPermissionSettingsRejectsInvalidGrants(t *testing.T) {
	settings := ForumPermissionSettings{
		ForumID:  uuid.New(),
		Creators: []ForumPermissionGrant{{SubjectType: PermissionSubjectUser}},
	}

	err := settings.Validate()

	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestForumPermissionSimulationRequestNormalizeUsesForum verifies request defaults.
func TestForumPermissionSimulationRequestNormalizeUsesForum(t *testing.T) {
	forumID := uuid.New()
	request := ForumPermissionSimulationRequest{Permission: " forums.view "}.Normalize(forumID)

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if request.Permission != "forums.view" || request.ObjectType != "forum" || request.ObjectID != forumID {
		t.Fatalf("Normalize() = %+v, want forum target", request)
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
	post := Post{
		ID:                  uuid.New(),
		ThreadID:            uuid.New(),
		ForumID:             uuid.New(),
		AuthorUserID:        uuid.New(),
		Sequence:            1,
		ContentDocumentJSON: []byte(`[]`),
		ContentText:         "hello",
	}.Normalize()
	err := post.Validate()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestThreadAndPostValidateAcceptValidContent verifies content entities accept valid state.
func TestThreadAndPostValidateAcceptValidContent(t *testing.T) {
	authorID := uuid.New()
	thread := Thread{
		ID:                     uuid.New(),
		ForumID:                uuid.New(),
		AuthorUserID:           authorID,
		OpenerPostID:           uuid.New(),
		LatestPostID:           uuid.New(),
		LatestPostAuthorUserID: authorID,
		Title:                  "Valid title",
		Slug:                   "valid-title",
	}.Normalize()
	if err := thread.Validate(); err != nil {
		t.Fatalf("Thread Validate() error = %v", err)
	}
	post := Post{
		ID:                  uuid.New(),
		ThreadID:            thread.ID,
		ForumID:             thread.ForumID,
		AuthorUserID:        authorID,
		Sequence:            1,
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "hello",
	}.Normalize()
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

// TestPostLikeValidateRequiresIdentity verifies like identity validation.
func TestPostLikeValidateRequiresIdentity(t *testing.T) {
	like := PostLike{ID: uuid.New(), PostID: uuid.New(), ThreadID: uuid.New(), ForumID: uuid.New(), UserID: uuid.New()}

	if err := like.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestThreadReadStateValidateRejectsAnonymousOrEmptySequence verifies read-state validation.
func TestThreadReadStateValidateRejectsAnonymousOrEmptySequence(t *testing.T) {
	state := ThreadReadState{ID: uuid.New(), ForumID: uuid.New(), ThreadID: uuid.New(), LastReadAt: time.Now().UTC()}

	err := state.Validate()
	if err == nil {
		t.Fatalf("Validate() error = nil, want validation error")
	}
}

// TestValidationForwardersCoverCompatibilityLayer verifies parent-domain helper wrappers.
func TestValidationForwardersCoverCompatibilityLayer(t *testing.T) {
	validations := [][]Violation{
		ValidateKey("key", "forum_news"),
		ValidateSlug("slug", "forum-news"),
		ValidateName("name", "News"),
		ValidateDescription("description", "Updates"),
		ValidateDisplayOrder("display_order", 0),
		ValidateCategoryStatus("status", CategoryStatusActive),
		ValidateForumKind("kind", ForumKindDiscussion),
		ValidateForumStatus("status", ForumStatusActive),
		ValidateThreadVisibilityMode("visibility", ThreadVisibilityAllThreads),
		ValidateThreadStatus("status", ThreadStatusOpen),
		ValidateStickyState("sticky", StickyStateNormal),
		ValidatePostStatus("status", PostStatusVisible),
		ValidateContentFormat("format", ContentFormatProseMirror),
		ValidateReferenceType("reference", ReferenceLink),
		ValidatePermissionSubjectType("subject", PermissionSubjectPublic),
		ValidateExternalURL("url", "https://example.com"),
		ValidateTitle("title", "Valid title"),
		ValidateContentDocument("document", json.RawMessage(`{"type":"doc"}`)),
		ValidateContentText("text", "hello"),
	}
	for _, violations := range validations {
		if len(violations) != 0 {
			t.Fatalf("expected wrapper validation to pass: %#v", violations)
		}
	}

	err := NewValidationError(AppendViolation(nil, "field", "message"))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("NewValidationError() = %v, want ErrInvalid", err)
	}
}
