package content

import (
	"encoding/json"

	"github.com/realmkit/rk-backend/module/forums/domain"
)

// contentRequest is a post content write request.
type contentRequest struct {
	ContentDocumentJSON json.RawMessage        `json:"content_document_json"` // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string                 `json:"content_text"`          // ContentText stores the content text value.
	ContentChecksum     string                 `json:"content_checksum"`      // ContentChecksum stores the content checksum value.
	EditReason          string                 `json:"edit_reason"`           // EditReason stores the edit reason value.
	References          []domain.PostReference `json:"references"`            // References stores the references value.
}

// threadCreateRequest is a thread creation request.
type threadCreateRequest struct {
	Title               string          `json:"title"`                 // Title stores the title value.
	Slug                domain.Slug     `json:"slug"`                  // Slug stores the slug value.
	ContentDocumentJSON json.RawMessage `json:"content_document_json"` // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string          `json:"content_text"`          // ContentText stores the content text value.
	ContentChecksum     string          `json:"content_checksum"`      // ContentChecksum stores the content checksum value.
}

// threadUpdateRequest is a thread title update request.
type threadUpdateRequest struct {
	Title string      `json:"title"` // Title stores the title value.
	Slug  domain.Slug `json:"slug"`  // Slug stores the slug value.
}

// threadCreateResponse returns created thread and opener post.
type threadCreateResponse struct {
	Thread domain.Thread `json:"thread"` // Thread stores the thread value.
	Post   domain.Post   `json:"post"`   // Post stores the post value.
}

// threadListResponse contains one thread page.
type threadListResponse struct {
	Items         []domain.Thread `json:"items"`                     // Items stores the items value.
	NextPageToken string          `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// postListResponse contains one post page.
type postListResponse struct {
	Items         []domain.Post `json:"items"`                     // Items stores the items value.
	NextPageToken string        `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// revisionListResponse contains one revision page.
type revisionListResponse struct {
	Items         []domain.PostRevision `json:"items"`                     // Items stores the items value.
	NextPageToken string                `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// latestPostListResponse contains latest-post widget rows.
type latestPostListResponse struct {
	Items         []domain.LatestPostSummary `json:"items"`                     // Items stores the items value.
	NextPageToken string                     `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// mostLikedPostListResponse contains most-liked widget rows.
type mostLikedPostListResponse struct {
	Items         []domain.MostLikedPost `json:"items"`                     // Items stores the items value.
	NextPageToken string                 `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// searchListResponse contains forum search rows.
type searchListResponse struct {
	Items         []domain.SearchResult `json:"items"`                     // Items stores the items value.
	NextPageToken string                `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// readThreadRequest marks a thread read through a sequence.
type readThreadRequest struct {
	LastReadPostSequence int64 `json:"last_read_post_sequence"` // LastReadPostSequence stores the last read post sequence value.
}
