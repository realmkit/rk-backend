// Package structure adapts forum structure and admin use cases to Fiber routes.
package structure

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// categoryRequest is a category write request.
type categoryRequest struct {
	Key          domain.Key            `json:"key"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	DisplayOrder int                   `json:"display_order"`
	Status       domain.CategoryStatus `json:"status"`
}

// reorderRequest is a display-order request.
type reorderRequest struct {
	Items []port.ReorderItem `json:"items"`
}

// categoryListResponse contains one category page.
type categoryListResponse struct {
	Items         []domain.ForumCategory `json:"items"`
	NextPageToken string                 `json:"next_page_token,omitempty"`
}

// forumRequest is a forum write request.
type forumRequest struct {
	CategoryID                    uuid.UUID                   `json:"category_id"`
	ParentForumID                 *uuid.UUID                  `json:"parent_forum_id"`
	Kind                          domain.ForumKind            `json:"kind"`
	Key                           domain.Key                  `json:"key"`
	Slug                          domain.Slug                 `json:"slug"`
	Name                          string                      `json:"name"`
	Description                   string                      `json:"description"`
	DisplayOrder                  int                         `json:"display_order"`
	ExternalURL                   string                      `json:"external_url"`
	IconAssetID                   *uuid.UUID                  `json:"icon_asset_id"`
	ThreadVisibilityMode          domain.ThreadVisibilityMode `json:"thread_visibility_mode"`
	MaxStickyThreads              int                         `json:"max_sticky_threads"`
	DefaultThreadStatus           domain.ThreadStatus         `json:"default_thread_status"`
	AuthorPostEditWindowSeconds   int                         `json:"author_post_edit_window_seconds"`
	AuthorPostDeleteWindowSeconds int                         `json:"author_post_delete_window_seconds"`
	Status                        domain.ForumStatus          `json:"status"`
}

// moveForumRequest is a forum move request.
type moveForumRequest struct {
	CategoryID    uuid.UUID  `json:"category_id"`
	ParentForumID *uuid.UUID `json:"parent_forum_id"`
	DisplayOrder  int        `json:"display_order"`
}

// forumListResponse contains one forum page.
type forumListResponse struct {
	Items         []domain.Forum `json:"items"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}
