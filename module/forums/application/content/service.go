// Package content implements thread, post, revision, and reference use cases.
package content

import (
	"context"

	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/transaction"
)

// Service manages forum content use cases.
type Service struct {
	forums       port.ForumRepository      // forums stores the forums value.
	threads      port.ThreadRepository     // threads stores the threads value.
	posts        port.PostRepository       // posts stores the posts value.
	assets       port.AssetResolver        // assets stores the assets value.
	authorizer   port.VisibilityAuthorizer // authorizer stores the authorizer value.
	restrictions port.RestrictionChecker   // restrictions stores the restrictions value.
	cache        port.ReadCache            // cache stores the cache value.
	transactions transaction.Runner        // transactions stores the transactions value.
	events       emitter.Publisher         // events stores the events value.
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

	// Restrictions checks punishment restrictions.
	Restrictions port.RestrictionChecker

	// Cache clears read models affected by writes.
	Cache port.ReadCache

	// Transactions runs transactional use cases.
	Transactions transaction.Runner

	// Events publishes forum content events.
	Events emitter.Publisher
}

// NewService creates a content service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:       deps.Forums,
		threads:      deps.Threads,
		posts:        deps.Posts,
		assets:       deps.Assets,
		authorizer:   deps.Authorizer,
		restrictions: deps.Restrictions,
		cache:        deps.Cache,
		transactions: deps.Transactions,
		events:       deps.Events,
	}
}

// clearTree clears cached trees when a cache is configured.
func (service Service) clearTree(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearTree(ctx)
}
