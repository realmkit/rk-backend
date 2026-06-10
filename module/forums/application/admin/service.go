// Package admin implements forum settings and permission configuration use cases.
package admin

import (
	"context"

	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/transaction"
)

// Service manages forum admin use cases.
type Service struct {
	forums       port.ForumRepository
	authorizer   port.VisibilityAuthorizer
	permissions  port.PermissionAdmin
	cache        port.ReadCache
	transactions transaction.Runner
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
}

// NewService creates an admin service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:       deps.Forums,
		authorizer:   deps.Authorizer,
		permissions:  deps.Permissions,
		cache:        deps.Cache,
		transactions: deps.Transactions,
	}
}

// clearReadCache clears every forum read cache.
func (service Service) clearReadCache(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearAll(ctx)
}
