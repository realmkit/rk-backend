// Package operations implements forum search, repair, cache, and view flush use cases.
package operations

import "github.com/niflaot/gamehub-go/module/forums/port"

// Service manages forum operational use cases.
type Service struct {
	forums     port.ForumRepository
	operations port.OperationsRepository
	authorizer port.VisibilityAuthorizer
	cache      port.ReadCache
}

// Dependencies contains operations service dependencies.
type Dependencies struct {
	// Forums stores forums.
	Forums port.ForumRepository

	// Operations runs search, repair, and counter flushes.
	Operations port.OperationsRepository

	// Authorizer checks forum visibility.
	Authorizer port.VisibilityAuthorizer

	// Cache stores operational read buffers and caches.
	Cache port.ReadCache
}

// NewService creates an operations service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:     deps.Forums,
		operations: deps.Operations,
		authorizer: deps.Authorizer,
		cache:      deps.Cache,
	}
}
