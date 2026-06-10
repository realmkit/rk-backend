package content

import (
	"encoding/json"

	"github.com/niflaot/gamehub-go/module/forums/domain"
)

// contentRequest is a post content write request.
type contentRequest struct {
	ContentDocumentJSON json.RawMessage        `json:"content_document_json"`
	ContentText         string                 `json:"content_text"`
	ContentChecksum     string                 `json:"content_checksum"`
	EditReason          string                 `json:"edit_reason"`
	References          []domain.PostReference `json:"references"`
}

// threadCreateRequest is a thread creation request.
type threadCreateRequest struct {
	Title               string          `json:"title"`
	Slug                domain.Slug     `json:"slug"`
	ContentDocumentJSON json.RawMessage `json:"content_document_json"`
	ContentText         string          `json:"content_text"`
	ContentChecksum     string          `json:"content_checksum"`
}

// threadUpdateRequest is a thread title update request.
type threadUpdateRequest struct {
	Title string      `json:"title"`
	Slug  domain.Slug `json:"slug"`
}

// threadCreateResponse returns created thread and opener post.
type threadCreateResponse struct {
	Thread domain.Thread `json:"thread"`
	Post   domain.Post   `json:"post"`
}

// threadListResponse contains one thread page.
type threadListResponse struct {
	Items         []domain.Thread `json:"items"`
	NextPageToken string          `json:"next_page_token,omitempty"`
}

// postListResponse contains one post page.
type postListResponse struct {
	Items         []domain.Post `json:"items"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

// revisionListResponse contains one revision page.
type revisionListResponse struct {
	Items         []domain.PostRevision `json:"items"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

// latestPostListResponse contains latest-post widget rows.
type latestPostListResponse struct {
	Items         []domain.LatestPostSummary `json:"items"`
	NextPageToken string                     `json:"next_page_token,omitempty"`
}

// mostLikedPostListResponse contains most-liked widget rows.
type mostLikedPostListResponse struct {
	Items         []domain.MostLikedPost `json:"items"`
	NextPageToken string                 `json:"next_page_token,omitempty"`
}

// searchListResponse contains forum search rows.
type searchListResponse struct {
	Items         []domain.SearchResult `json:"items"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

// readThreadRequest marks a thread read through a sequence.
type readThreadRequest struct {
	LastReadPostSequence int64 `json:"last_read_post_sequence"`
}
