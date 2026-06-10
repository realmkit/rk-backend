// Package content implements thread, post, revision, and reference use cases.
package content

import (
	"context"

	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/transaction"
)

// Service manages forum content use cases.
type Service struct {
	forums       port.ForumRepository
	threads      port.ThreadRepository
	posts        port.PostRepository
	assets       port.AssetResolver
	authorizer   port.VisibilityAuthorizer
	cache        port.ReadCache
	transactions transaction.Runner
}

// Dependencies contains content service dependencies.
type Dependencies struct {
	// Forums stores forums.
	Forums port.ForumRepository

	// Threads stores threads.
	Threads port.ThreadRepository

	// Posts stores posts, revisions, and references.
	Posts port.PostRepository

	// Assets resolves attachment assets.
	Assets port.AssetResolver

	// Authorizer checks forum permissions.
	Authorizer port.VisibilityAuthorizer

	// Cache clears read models affected by writes.
	Cache port.ReadCache

	// Transactions runs transactional use cases.
	Transactions transaction.Runner
}

// NewService creates a content service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:       deps.Forums,
		threads:      deps.Threads,
		posts:        deps.Posts,
		assets:       deps.Assets,
		authorizer:   deps.Authorizer,
		cache:        deps.Cache,
		transactions: deps.Transactions,
	}
}

// clearTree clears cached trees when a cache is configured.
func (service Service) clearTree(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearTree(ctx)
}
