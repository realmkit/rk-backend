// Package structure adapts forum structure and admin use cases to Fiber routes.
package structure

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// categoryRequest is a category write request.
type categoryRequest struct {
	Key          domain.Key            `json:"key"`           // Key stores the key value.
	Name         string                `json:"name"`          // Name stores the name value.
	Description  string                `json:"description"`   // Description stores the description value.
	DisplayOrder int                   `json:"display_order"` // DisplayOrder stores the display order value.
	Status       domain.CategoryStatus `json:"status"`        // Status stores the status value.
}

// reorderRequest is a display-order request.
type reorderRequest struct {
	Items []port.ReorderItem `json:"items"` // Items stores the items value.
}

// categoryListResponse contains one category page.
type categoryListResponse struct {
	Items         []domain.ForumCategory `json:"items"`                     // Items stores the items value.
	NextPageToken string                 `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// forumRequest is a forum write request.
type forumRequest struct {
	CategoryID                    uuid.UUID                   `json:"category_id"`                       // CategoryID stores the category i d value.
	ParentForumID                 *uuid.UUID                  `json:"parent_forum_id"`                   // ParentForumID stores the parent forum i d value.
	Kind                          domain.ForumKind            `json:"kind"`                              // Kind stores the kind value.
	Key                           domain.Key                  `json:"key"`                               // Key stores the key value.
	Slug                          domain.Slug                 `json:"slug"`                              // Slug stores the slug value.
	Name                          string                      `json:"name"`                              // Name stores the name value.
	Description                   string                      `json:"description"`                       // Description stores the description value.
	DisplayOrder                  int                         `json:"display_order"`                     // DisplayOrder stores the display order value.
	ExternalURL                   string                      `json:"external_url"`                      // ExternalURL stores the external u r l value.
	IconAssetID                   *uuid.UUID                  `json:"icon_asset_id"`                     // IconAssetID stores the icon asset i d value.
	ThreadVisibilityMode          domain.ThreadVisibilityMode `json:"thread_visibility_mode"`            // ThreadVisibilityMode stores the thread visibility mode value.
	MaxStickyThreads              int                         `json:"max_sticky_threads"`                // MaxStickyThreads stores the max sticky threads value.
	DefaultThreadStatus           domain.ThreadStatus         `json:"default_thread_status"`             // DefaultThreadStatus stores the default thread status value.
	AuthorPostEditWindowSeconds   int                         `json:"author_post_edit_window_seconds"`   // AuthorPostEditWindowSeconds stores the author post edit window seconds value.
	AuthorPostDeleteWindowSeconds int                         `json:"author_post_delete_window_seconds"` // AuthorPostDeleteWindowSeconds stores the author post delete window seconds value.
	Status                        domain.ForumStatus          `json:"status"`                            // Status stores the status value.
}

// moveForumRequest is a forum move request.
type moveForumRequest struct {
	CategoryID    uuid.UUID  `json:"category_id"`     // CategoryID stores the category i d value.
	ParentForumID *uuid.UUID `json:"parent_forum_id"` // ParentForumID stores the parent forum i d value.
	DisplayOrder  int        `json:"display_order"`   // DisplayOrder stores the display order value.
}

// forumListResponse contains one forum page.
type forumListResponse struct {
	Items         []domain.Forum `json:"items"`                     // Items stores the items value.
	NextPageToken string         `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}
