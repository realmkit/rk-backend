package domain

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Forum is a discussion board, container, or utility link.
type Forum struct {
	// ID is the forum identifier.
	ID uuid.UUID `json:"id"`

	// CategoryID is the parent category.
	CategoryID uuid.UUID `json:"category_id"`

	// ParentForumID is the optional parent forum.
	ParentForumID *uuid.UUID `json:"parent_forum_id,omitempty"`

	// Kind is the forum structural kind.
	Kind ForumKind `json:"kind"`

	// Key is the stable forum key.
	Key Key `json:"key"`

	// Slug is the URL slug.
	Slug Slug `json:"slug"`

	// Name is the display name.
	Name string `json:"name"`

	// Description explains the forum.
	Description string `json:"description"`

	// DisplayOrder controls forum ordering among siblings.
	DisplayOrder int `json:"display_order"`

	// Path is the materialized tree path.
	Path string `json:"path"`

	// Depth is the tree depth.
	Depth int `json:"depth"`

	// ExternalURL is the target URL for link forums.
	ExternalURL string `json:"external_url"`

	// IconAssetID is the optional icon asset.
	IconAssetID *uuid.UUID `json:"icon_asset_id,omitempty"`

	// ThreadVisibilityMode controls thread list filtering.
	ThreadVisibilityMode ThreadVisibilityMode `json:"thread_visibility_mode"`

	// MaxStickyThreads limits sticky threads in this forum.
	MaxStickyThreads int `json:"max_sticky_threads"`

	// DefaultThreadStatus is the initial thread status.
	DefaultThreadStatus ThreadStatus `json:"default_thread_status"`

	// AuthorPostEditWindowSeconds is the author self-edit window in seconds.
	AuthorPostEditWindowSeconds int `json:"author_post_edit_window_seconds"`

	// AuthorPostDeleteWindowSeconds is the author self-delete window in seconds.
	AuthorPostDeleteWindowSeconds int `json:"author_post_delete_window_seconds"`

	// Status is the forum lifecycle state.
	Status ForumStatus `json:"status"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized forum copy.
func (forum Forum) Normalize() Forum {
	forum.Key = Key(strings.TrimSpace(string(forum.Key)))
	forum.Slug = Slug(strings.TrimSpace(string(forum.Slug)))
	forum.Name = strings.TrimSpace(forum.Name)
	forum.Description = strings.TrimSpace(forum.Description)
	forum.ExternalURL = strings.TrimSpace(forum.ExternalURL)
	if forum.Kind == "" {
		forum.Kind = ForumKindDiscussion
	}
	if forum.ThreadVisibilityMode == "" {
		forum.ThreadVisibilityMode = ThreadVisibilityAllThreads
	}
	if forum.DefaultThreadStatus == "" {
		forum.DefaultThreadStatus = ThreadStatusOpen
	}
	if forum.AuthorPostEditWindowSeconds == 0 {
		forum.AuthorPostEditWindowSeconds = DefaultAuthorPostEditWindowSeconds
	}
	if forum.AuthorPostDeleteWindowSeconds == 0 {
		forum.AuthorPostDeleteWindowSeconds = DefaultAuthorPostDeleteWindowSeconds
	}
	if forum.Status == "" {
		forum.Status = ForumStatusActive
	}
	if forum.Version == 0 {
		forum.Version = 1
	}
	return forum
}

// Validate validates forum fields.
func (forum Forum) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateKey("key", forum.Key)...)
	violations = append(violations, ValidateSlug("slug", forum.Slug)...)
	violations = append(violations, ValidateName("name", forum.Name)...)
	violations = append(violations, ValidateDescription("description", forum.Description)...)
	violations = append(violations, ValidateDisplayOrder("display_order", forum.DisplayOrder)...)
	violations = append(violations, ValidateForumKind("kind", forum.Kind)...)
	violations = append(violations, ValidateForumStatus("status", forum.Status)...)
	violations = append(violations, ValidateThreadVisibilityMode("thread_visibility_mode", forum.ThreadVisibilityMode)...)
	violations = append(violations, ValidateThreadStatus("default_thread_status", forum.DefaultThreadStatus)...)
	if forum.CategoryID == uuid.Nil {
		violations = AppendViolation(violations, "category_id", "is required")
	}
	if forum.Depth < 0 || forum.Depth > 5 {
		violations = AppendViolation(violations, "depth", "must be between 0 and 5")
	}
	if forum.MaxStickyThreads < 0 {
		violations = AppendViolation(violations, "max_sticky_threads", "must be zero or greater")
	}
	if forum.AuthorPostEditWindowSeconds < -1 {
		violations = AppendViolation(violations, "author_post_edit_window_seconds", "must be -1 or greater")
	}
	if forum.AuthorPostDeleteWindowSeconds < -1 {
		violations = AppendViolation(violations, "author_post_delete_window_seconds", "must be -1 or greater")
	}
	if forum.Kind == ForumKindLink {
		violations = append(violations, ValidateExternalURL("external_url", forum.ExternalURL)...)
	}
	if forum.Kind != ForumKindLink && forum.ExternalURL != "" {
		violations = AppendViolation(violations, "external_url", "is only supported by link forums")
	}
	if !validPath(forum.Path, forum.ID) {
		violations = AppendViolation(violations, "path", "must be a materialized path ending with the forum id")
	}
	return NewValidationError(violations)
}

// Discussion reports whether the forum can contain threads.
func (forum Forum) Discussion() bool {
	return forum.Kind == ForumKindDiscussion
}

// Settings returns the admin settings view for the forum.
func (forum Forum) Settings() ForumSettings {
	return ForumSettings{ForumID: forum.ID, Kind: forum.Kind, ExternalURL: forum.ExternalURL, ThreadVisibilityMode: forum.ThreadVisibilityMode, MaxStickyThreads: forum.MaxStickyThreads, DefaultThreadStatus: forum.DefaultThreadStatus, AuthorPostEditWindowSeconds: forum.AuthorPostEditWindowSeconds, AuthorPostDeleteWindowSeconds: forum.AuthorPostDeleteWindowSeconds, Version: forum.Version, UpdatedAt: forum.UpdatedAt}
}

// ForumSettings contains admin-editable forum behavior settings.
type ForumSettings struct {
	// ForumID is the configured forum.
	ForumID uuid.UUID `json:"forum_id"`

	// Kind is the forum structural kind.
	Kind ForumKind `json:"kind"`

	// ExternalURL is required for link forums.
	ExternalURL string `json:"external_url"`

	// ThreadVisibilityMode shapes thread-list SQL for normal readers.
	ThreadVisibilityMode ThreadVisibilityMode `json:"thread_visibility_mode"`

	// MaxStickyThreads limits sticky threads.
	MaxStickyThreads int `json:"max_sticky_threads"`

	// DefaultThreadStatus is applied to new threads.
	DefaultThreadStatus ThreadStatus `json:"default_thread_status"`

	// AuthorPostEditWindowSeconds is the self-edit window, or -1 when disabled.
	AuthorPostEditWindowSeconds int `json:"author_post_edit_window_seconds"`

	// AuthorPostDeleteWindowSeconds is the self-delete window, or -1 when disabled.
	AuthorPostDeleteWindowSeconds int `json:"author_post_delete_window_seconds"`

	// Version is the forum optimistic concurrency version.
	Version uint64 `json:"version"`

	// UpdatedAt is the forum settings update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized settings copy.
func (settings ForumSettings) Normalize() ForumSettings {
	settings.ExternalURL = strings.TrimSpace(settings.ExternalURL)
	if settings.Kind == "" {
		settings.Kind = ForumKindDiscussion
	}
	if settings.ThreadVisibilityMode == "" {
		settings.ThreadVisibilityMode = ThreadVisibilityAllThreads
	}
	if settings.DefaultThreadStatus == "" {
		settings.DefaultThreadStatus = ThreadStatusOpen
	}
	if settings.AuthorPostEditWindowSeconds == 0 {
		settings.AuthorPostEditWindowSeconds = DefaultAuthorPostEditWindowSeconds
	}
	if settings.AuthorPostDeleteWindowSeconds == 0 {
		settings.AuthorPostDeleteWindowSeconds = DefaultAuthorPostDeleteWindowSeconds
	}
	return settings
}

// Validate validates forum settings.
func (settings ForumSettings) Validate() error {
	var violations []Violation
	if settings.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	violations = append(violations, ValidateForumKind("kind", settings.Kind)...)
	violations = append(violations, ValidateThreadVisibilityMode("thread_visibility_mode", settings.ThreadVisibilityMode)...)
	violations = append(violations, ValidateThreadStatus("default_thread_status", settings.DefaultThreadStatus)...)
	if settings.MaxStickyThreads < 0 {
		violations = AppendViolation(violations, "max_sticky_threads", "must be zero or greater")
	}
	if settings.AuthorPostEditWindowSeconds < -1 {
		violations = AppendViolation(violations, "author_post_edit_window_seconds", "must be -1 or greater")
	}
	if settings.AuthorPostDeleteWindowSeconds < -1 {
		violations = AppendViolation(violations, "author_post_delete_window_seconds", "must be -1 or greater")
	}
	if settings.Kind == ForumKindLink {
		violations = append(violations, ValidateExternalURL("external_url", settings.ExternalURL)...)
	}
	if settings.Kind != ForumKindLink && settings.ExternalURL != "" {
		violations = AppendViolation(violations, "external_url", "is only supported by link forums")
	}
	return NewValidationError(violations)
}

// ForumPermissionGrant grants one forum relation to one subject.
type ForumPermissionGrant struct {
	// SubjectType is public, authenticated, user, or group.
	SubjectType PermissionSubjectType `json:"subject_type"`

	// SubjectID is the concrete subject identifier.
	SubjectID uuid.UUID `json:"subject_id"`

	// SubjectRelation is usually member for group subjects.
	SubjectRelation string `json:"subject_relation,omitempty"`
}

// Normalize returns a normalized permission grant.
func (grant ForumPermissionGrant) Normalize() ForumPermissionGrant {
	grant.SubjectType = PermissionSubjectType(strings.TrimSpace(string(grant.SubjectType)))
	grant.SubjectRelation = strings.TrimSpace(grant.SubjectRelation)
	switch grant.SubjectType {
	case PermissionSubjectPublic:
		grant.SubjectID = PublicPermissionSubjectID()
	case PermissionSubjectAuthenticated:
		grant.SubjectID = AuthenticatedPermissionSubjectID()
	case PermissionSubjectGroup:
		if grant.SubjectRelation == "" {
			grant.SubjectRelation = "member"
		}
	}
	return grant
}

// Validate validates a forum permission grant.
func (grant ForumPermissionGrant) Validate(field string) []Violation {
	var violations []Violation
	violations = append(violations, ValidatePermissionSubjectType(field+".subject_type", grant.SubjectType)...)
	switch grant.SubjectType {
	case PermissionSubjectPublic:
		if grant.SubjectID != PublicPermissionSubjectID() {
			violations = AppendViolation(violations, field+".subject_id", "must use the public reserved identifier")
		}
		if grant.SubjectRelation != "" {
			violations = AppendViolation(violations, field+".subject_relation", "must be empty for public grants")
		}
	case PermissionSubjectAuthenticated:
		if grant.SubjectID != AuthenticatedPermissionSubjectID() {
			violations = AppendViolation(violations, field+".subject_id", "must use the authenticated reserved identifier")
		}
		if grant.SubjectRelation != "" {
			violations = AppendViolation(violations, field+".subject_relation", "must be empty for authenticated grants")
		}
	case PermissionSubjectUser:
		if grant.SubjectID == uuid.Nil {
			violations = AppendViolation(violations, field+".subject_id", "is required")
		}
		if grant.SubjectRelation != "" {
			violations = AppendViolation(violations, field+".subject_relation", "must be empty for user grants")
		}
	case PermissionSubjectGroup:
		if grant.SubjectID == uuid.Nil {
			violations = AppendViolation(violations, field+".subject_id", "is required")
		}
		if grant.SubjectRelation != "member" {
			violations = AppendViolation(violations, field+".subject_relation", "must be member for group grants")
		}
	}
	return violations
}

// ForumPermissionSettings contains forum relation grants.
type ForumPermissionSettings struct {
	// ForumID is the configured forum.
	ForumID uuid.UUID `json:"forum_id"`

	// Viewers can view the forum.
	Viewers []ForumPermissionGrant `json:"viewers"`

	// Creators can create threads.
	Creators []ForumPermissionGrant `json:"creators"`

	// Replyers can reply to threads.
	Replyers []ForumPermissionGrant `json:"replyers"`

	// Likers can like posts.
	Likers []ForumPermissionGrant `json:"likers"`

	// Moderators can moderate threads, posts, and sticky state.
	Moderators []ForumPermissionGrant `json:"moderators"`

	// Managers can manage forum settings and permissions.
	Managers []ForumPermissionGrant `json:"managers"`
}

// Normalize returns normalized permission settings.
func (settings ForumPermissionSettings) Normalize() ForumPermissionSettings {
	settings.Viewers = normalizePermissionGrants(settings.Viewers)
	settings.Creators = normalizePermissionGrants(settings.Creators)
	settings.Replyers = normalizePermissionGrants(settings.Replyers)
	settings.Likers = normalizePermissionGrants(settings.Likers)
	settings.Moderators = normalizePermissionGrants(settings.Moderators)
	settings.Managers = normalizePermissionGrants(settings.Managers)
	return settings
}

// Validate validates permission settings.
func (settings ForumPermissionSettings) Validate() error {
	var violations []Violation
	if settings.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	violations = append(violations, validatePermissionGrants("viewers", settings.Viewers)...)
	violations = append(violations, validatePermissionGrants("creators", settings.Creators)...)
	violations = append(violations, validatePermissionGrants("replyers", settings.Replyers)...)
	violations = append(violations, validatePermissionGrants("likers", settings.Likers)...)
	violations = append(violations, validatePermissionGrants("moderators", settings.Moderators)...)
	violations = append(violations, validatePermissionGrants("managers", settings.Managers)...)
	return NewValidationError(violations)
}

// ForumPermissionSimulationRequest requests an explainable forum permission decision.
type ForumPermissionSimulationRequest struct {
	// ActorUserID is the selected actor; nil means anonymous.
	ActorUserID uuid.UUID `json:"actor_user_id,omitempty"`

	// Permission is the permission name to evaluate.
	Permission string `json:"permission"`

	// ObjectType is the object type to explain.
	ObjectType string `json:"object_type"`

	// ObjectID is the object identifier to explain.
	ObjectID uuid.UUID `json:"object_id,omitempty"`
}

// Normalize returns a normalized simulation request.
func (request ForumPermissionSimulationRequest) Normalize(forumID uuid.UUID) ForumPermissionSimulationRequest {
	request.Permission = strings.TrimSpace(request.Permission)
	request.ObjectType = strings.TrimSpace(request.ObjectType)
	if request.ObjectType == "" {
		request.ObjectType = "forum"
	}
	if request.ObjectID == uuid.Nil {
		request.ObjectID = forumID
	}
	return request
}

// Validate validates a simulation request.
func (request ForumPermissionSimulationRequest) Validate() error {
	var violations []Violation
	if request.Permission == "" {
		violations = AppendViolation(violations, "permission", "is required")
	}
	if request.ObjectType == "" {
		violations = AppendViolation(violations, "object_type", "is required")
	}
	if request.ObjectID == uuid.Nil {
		violations = AppendViolation(violations, "object_id", "is required")
	}
	return NewValidationError(violations)
}

// ForumPermissionSimulationResult explains a forum permission decision.
type ForumPermissionSimulationResult struct {
	// Allowed reports whether the actor is allowed.
	Allowed bool `json:"allowed"`

	// Reason explains the decision.
	Reason string `json:"reason"`

	// Permission is the evaluated permission.
	Permission string `json:"permission"`

	// ObjectType is the evaluated object type.
	ObjectType string `json:"object_type"`

	// ObjectID is the evaluated object identifier.
	ObjectID uuid.UUID `json:"object_id"`

	// MatchedRelation is the relation that allowed the request.
	MatchedRelation string `json:"matched_relation,omitempty"`

	// CheckedRelations are the relations considered.
	CheckedRelations []string `json:"checked_relations"`
}

// PublicPermissionSubjectID returns the public grant subject identifier.
func PublicPermissionSubjectID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// AuthenticatedPermissionSubjectID returns the authenticated grant subject identifier.
func AuthenticatedPermissionSubjectID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000002")
}

// normalizePermissionGrants normalizes grants.
func normalizePermissionGrants(grants []ForumPermissionGrant) []ForumPermissionGrant {
	normalized := make([]ForumPermissionGrant, 0, len(grants))
	for _, grant := range grants {
		normalized = append(normalized, grant.Normalize())
	}
	return normalized
}

// validatePermissionGrants validates grants at field.
func validatePermissionGrants(field string, grants []ForumPermissionGrant) []Violation {
	var violations []Violation
	for index, grant := range grants {
		violations = append(violations, grant.Validate(field+"["+strconv.Itoa(index)+"]")...)
	}
	return violations
}

// validPath reports whether path is a valid materialized forum path.
func validPath(path string, id uuid.UUID) bool {
	if id == uuid.Nil {
		return true
	}
	expected := "/" + id.String() + "/"
	return strings.HasSuffix(path, expected) && strings.HasPrefix(path, "/")
}
