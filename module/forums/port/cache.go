package port

import (
	"context"
	"time"

	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// ReadCache caches visible forum read paths.
type ReadCache interface {
	// GetTree returns a cached tree when present.
	GetTree(ctx context.Context, key string) (domain.ForumTree, bool, error)

	// SetTree stores a tree for ttl.
	SetTree(ctx context.Context, key string, tree domain.ForumTree, ttl time.Duration) error

	// ClearTree removes forum tree cache entries.
	ClearTree(ctx context.Context) error

	// GetLatestPosts returns a cached latest-post page when present.
	GetLatestPosts(ctx context.Context, key string) (pagination.Result[domain.LatestPostSummary], bool, error)

	// SetLatestPosts stores a latest-post page for ttl.
	SetLatestPosts(ctx context.Context, key string, result pagination.Result[domain.LatestPostSummary], ttl time.Duration) error

	// ClearLatestPosts removes latest-post cache entries.
	ClearLatestPosts(ctx context.Context) error

	// GetMostLikedPosts returns a cached most-liked page when present.
	GetMostLikedPosts(ctx context.Context, key string) (pagination.Result[domain.MostLikedPost], bool, error)

	// SetMostLikedPosts stores a most-liked page for ttl.
	SetMostLikedPosts(ctx context.Context, key string, result pagination.Result[domain.MostLikedPost], ttl time.Duration) error

	// ClearMostLikedPosts removes most-liked cache entries.
	ClearMostLikedPosts(ctx context.Context) error
}
