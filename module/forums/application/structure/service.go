// Package structure implements category, forum, and visible-tree use cases.
package structure

import (
	"time"

	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/transaction"
)

// treeCacheTTL is the visible forum tree cache lifetime.
const treeCacheTTL = 30 * time.Second

// Service manages forum structure use cases.
type Service struct {
	categories   port.CategoryRepository
	forums       port.ForumRepository
	authorizer   port.VisibilityAuthorizer
	cache        port.ReadCache
	transactions transaction.Runner
	events       emitter.Publisher
}

// Dependencies contains structure service dependencies.
type Dependencies struct {
	// Categories stores categories.
	Categories port.CategoryRepository

	// Forums stores forums.
	Forums port.ForumRepository

	// Authorizer checks forum permissions.
	Authorizer port.VisibilityAuthorizer

	// Cache caches visible trees.
	Cache port.ReadCache

	// Transactions runs transactional use cases.
	Transactions transaction.Runner

	// Events publishes forum structure events.
	Events emitter.Publisher
}

// NewService creates a structure service.
func NewService(deps Dependencies) Service {
	return Service{
		categories:   deps.Categories,
		forums:       deps.Forums,
		authorizer:   deps.Authorizer,
		cache:        deps.Cache,
		transactions: deps.Transactions,
		events:       deps.Events,
	}
}
