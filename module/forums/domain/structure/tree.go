package structure

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ForumStats stores denormalized forum counters and latest post summary.
type ForumStats struct {
	// ForumID is the forum identifier.
	ForumID uuid.UUID `json:"forum_id"`

	// ThreadCount is the total active thread count.
	ThreadCount int64 `json:"thread_count"`

	// VisibleThreadCount is the active visible thread count.
	VisibleThreadCount int64 `json:"visible_thread_count"`

	// PostCount is the total active post count.
	PostCount int64 `json:"post_count"`

	// VisiblePostCount is the active visible post count.
	VisiblePostCount int64 `json:"visible_post_count"`

	// LatestThreadID is the latest visible thread.
	LatestThreadID *uuid.UUID `json:"latest_thread_id,omitempty"`

	// LatestPostID is the latest visible post.
	LatestPostID *uuid.UUID `json:"latest_post_id,omitempty"`

	// LatestPostAuthorUserID is the latest post author.
	LatestPostAuthorUserID *uuid.UUID `json:"latest_post_author_user_id,omitempty"`

	// LatestPostAt is the latest visible post time.
	LatestPostAt *time.Time `json:"latest_post_at,omitempty"`

	// UpdatedAt is the last stats update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// ForumNode is one visible forum tree node.
type ForumNode struct {
	// Forum is the visible forum.
	Forum Forum `json:"forum"`

	// Stats contains denormalized counters.
	Stats ForumStats `json:"stats"`

	// Children contains visible child forums.
	Children []ForumNode `json:"children"`
}

// CategoryNode is one visible category with forums.
type CategoryNode struct {
	// Category is the visible category.
	Category ForumCategory `json:"category"`

	// Forums contains visible top-level forums.
	Forums []ForumNode `json:"forums"`
}

// ForumTree is the forum home tree response.
type ForumTree struct {
	// Categories contains visible categories.
	Categories []CategoryNode `json:"categories"`
}

func validPath(path string, id uuid.UUID) bool {
	if id == uuid.Nil {
		return true
	}
	expected := "/" + id.String() + "/"
	return strings.HasSuffix(path, expected) && strings.HasPrefix(path, "/")
}
