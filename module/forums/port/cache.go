package port

import (
	"context"
	"time"

	"github.com/niflaot/gamehub-go/module/forums/domain"
)

// TreeCache caches visible forum trees.
type TreeCache interface {
	// GetTree returns a cached tree when present.
	GetTree(ctx context.Context, key string) (domain.ForumTree, bool, error)

	// SetTree stores a tree for ttl.
	SetTree(ctx context.Context, key string, tree domain.ForumTree, ttl time.Duration) error

	// ClearTree removes forum tree cache entries.
	ClearTree(ctx context.Context) error
}
