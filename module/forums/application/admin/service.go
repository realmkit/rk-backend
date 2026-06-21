// Package admin implements forum settings and permission configuration use cases.
package admin

import (
	"context"

	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/transaction"
)

// Service manages forum admin use cases.
type Service struct {
	forums       port.ForumRepository      // forums stores the forums value.
	authorizer   port.VisibilityAuthorizer // authorizer stores the authorizer value.
	permissions  port.PermissionAdmin      // permissions stores the permissions value.
	cache        port.ReadCache            // cache stores the cache value.
	transactions transaction.Runner        // transactions stores the transactions value.
	events       emitter.Publisher         // events stores the events value.
}

// Dependencies contains admin service dependencies.
type Dependencies struct {
	// Forums stores forums.
	Forums port.ForumRepository

	// Authorizer checks forum permissions.
	Authorizer port.VisibilityAuthorizer

	// Permissions manages forum permission configuration.
	Permissions port.PermissionAdmin

	// Cache clears affected forum read caches.
	Cache port.ReadCache

	// Transactions runs transactional use cases.
	Transactions transaction.Runner

	// Events publishes forum admin events.
	Events emitter.Publisher
}

// NewService creates an admin service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:       deps.Forums,
		authorizer:   deps.Authorizer,
		permissions:  deps.Permissions,
		cache:        deps.Cache,
		transactions: deps.Transactions,
		events:       deps.Events,
	}
}

// clearReadCache clears every forum read cache.
func (service Service) clearReadCache(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearAll(ctx)
}
