package content

import (
	"time"

	"github.com/google/uuid"
)

// SearchResult is one forum search result row.
type SearchResult struct {
	// Type identifies whether the result is a thread or post.
	Type string `json:"type"`

	// ForumID is the containing forum.
	ForumID uuid.UUID `json:"forum_id"`

	// ThreadID is the containing thread.
	ThreadID uuid.UUID `json:"thread_id"`

	// PostID is set for post results.
	PostID *uuid.UUID `json:"post_id,omitempty"`

	// Title is the thread title.
	Title string `json:"title"`

	// Slug is the thread slug.
	Slug Slug `json:"slug"`

	// Excerpt is matched text for preview.
	Excerpt string `json:"excerpt"`

	// AuthorUserID is the result author.
	AuthorUserID uuid.UUID `json:"author_user_id"`

	// CreatedAt is the result creation time.
	CreatedAt time.Time `json:"created_at"`
}

// CounterDrift describes one mismatched forum counter.
type CounterDrift struct {
	// ObjectType identifies the counter owner.
	ObjectType string `json:"object_type"`

	// ObjectID is the counter owner ID.
	ObjectID uuid.UUID `json:"object_id"`

	// Field is the mismatched counter field.
	Field string `json:"field"`

	// Expected is the recalculated value.
	Expected int64 `json:"expected"`

	// Actual is the stored value.
	Actual int64 `json:"actual"`
}

// CounterDriftReport is the result of counter verification or rebuild.
type CounterDriftReport struct {
	// Mismatches are detected counter drift rows.
	Mismatches []CounterDrift `json:"mismatches"`

	// Repaired reports whether rows were updated.
	Repaired bool `json:"repaired"`
}
