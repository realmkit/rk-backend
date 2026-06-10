// Package interaction implements likes, latest posts, most-liked posts, and read-state use cases.
package interaction

import (
	"time"

	"github.com/niflaot/gamehub-go/module/forums/port"
)

// widgetCacheTTL is the forum widget cache lifetime.
const widgetCacheTTL = 20 * time.Second

// Service manages forum interaction use cases.
type Service struct {
	forums       port.ForumRepository
	threads      port.ThreadRepository
	posts        port.PostRepository
	interactions port.InteractionRepository
	authorizer   port.VisibilityAuthorizer
	cache        port.ReadCache
}

// Dependencies contains interaction service dependencies.
type Dependencies struct {
	// Forums stores forums.
	Forums port.ForumRepository

	// Threads stores threads.
	Threads port.ThreadRepository

	// Posts stores posts.
	Posts port.PostRepository

	// Interactions stores likes, widgets, and read state.
	Interactions port.InteractionRepository

	// Authorizer checks forum permissions.
	Authorizer port.VisibilityAuthorizer

	// Cache caches widget reads.
	Cache port.ReadCache
}

// NewService creates an interaction service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:       deps.Forums,
		threads:      deps.Threads,
		posts:        deps.Posts,
		interactions: deps.Interactions,
		authorizer:   deps.Authorizer,
		cache:        deps.Cache,
	}
}
