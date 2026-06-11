// Package domain contains forum entities, value objects, and validation rules.
package domain

import contentmodel "github.com/realmkit/rk-backend/module/forums/domain/content"

// Thread is a forum conversation timeline.
type Thread = contentmodel.Thread

// Post is one message in a thread timeline.
type Post = contentmodel.Post

// PostRevision stores content before an edit.
type PostRevision = contentmodel.PostRevision

// PostReference stores structured relationships extracted from a post.
type PostReference = contentmodel.PostReference

// PostLike records one active user like for one post.
type PostLike = contentmodel.PostLike

// PostLikeSummary describes the current user's like state for a post.
type PostLikeSummary = contentmodel.PostLikeSummary

// ThreadReadState stores how far a user has read in a thread.
type ThreadReadState = contentmodel.ThreadReadState

// LatestPostSummary is a compact latest-post widget row.
type LatestPostSummary = contentmodel.LatestPostSummary

// MostLikedPost is a compact most-liked widget row.
type MostLikedPost = contentmodel.MostLikedPost

// UnreadSummary describes unread thread state for visible forums.
type UnreadSummary = contentmodel.UnreadSummary

// ForumUnreadSummary describes unread state for one forum.
type ForumUnreadSummary = contentmodel.ForumUnreadSummary

// SearchResult is one forum search result row.
type SearchResult = contentmodel.SearchResult

// CounterDrift describes one mismatched forum counter.
type CounterDrift = contentmodel.CounterDrift

// CounterDriftReport is the result of counter verification or rebuild.
type CounterDriftReport = contentmodel.CounterDriftReport
