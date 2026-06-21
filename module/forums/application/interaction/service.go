// Package interaction implements likes, latest posts, most-liked posts, and read-state use cases.
package interaction

import (
	"time"

	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// widgetCacheTTL is the forum widget cache lifetime.
const widgetCacheTTL = 20 * time.Second

// Service manages forum interaction use cases.
type Service struct {
	forums       port.ForumRepository       // forums stores the forums value.
	threads      port.ThreadRepository      // threads stores the threads value.
	posts        port.PostRepository        // posts stores the posts value.
	interactions port.InteractionRepository // interactions stores the interactions value.
	authorizer   port.VisibilityAuthorizer  // authorizer stores the authorizer value.
	restrictions port.RestrictionChecker    // restrictions stores the restrictions value.
	cache        port.ReadCache             // cache stores the cache value.
	events       emitter.Publisher          // events stores the events value.
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

	// Restrictions checks punishment restrictions.
	Restrictions port.RestrictionChecker

	// Cache caches widget reads.
	Cache port.ReadCache

	// Events publishes forum interaction events.
	Events emitter.Publisher
}

// NewService creates an interaction service.
func NewService(deps Dependencies) Service {
	return Service{
		forums:       deps.Forums,
		threads:      deps.Threads,
		posts:        deps.Posts,
		interactions: deps.Interactions,
		authorizer:   deps.Authorizer,
		restrictions: deps.Restrictions,
		cache:        deps.Cache,
		events:       deps.Events,
	}
}
