// Package application composes forum use-case services.
package application

import (
	adminapp "github.com/niflaot/gamehub-go/module/forums/application/admin"
	contentapp "github.com/niflaot/gamehub-go/module/forums/application/content"
	interactionapp "github.com/niflaot/gamehub-go/module/forums/application/interaction"
	operationsapp "github.com/niflaot/gamehub-go/module/forums/application/operations"
	structureapp "github.com/niflaot/gamehub-go/module/forums/application/structure"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
	"github.com/niflaot/gamehub-go/pkg/transaction"
)

// Structure exposes structure use cases through the facade.
type Structure struct {
	structureapp.Service
}

// Content exposes content use cases through the facade.
type Content struct {
	contentapp.Service
}

// Interaction exposes interaction use cases through the facade.
type Interaction struct {
	interactionapp.Service
}

// Operations exposes operations use cases through the facade.
type Operations struct {
	operationsapp.Service
}

// Admin exposes admin use cases through the facade.
type Admin struct {
	adminapp.Service
}

// Service composes concern-owned forum use-case services.
type Service struct {
	Structure
	Content
	Interaction
	Operations
	Admin
}

// Dependencies contains forum service dependencies.
type Dependencies struct {
	// Categories stores categories.
	Categories port.CategoryRepository

	// Forums stores forums.
	Forums port.ForumRepository

	// Threads stores threads.
	Threads port.ThreadRepository

	// Posts stores posts.
	Posts port.PostRepository

	// Interactions stores likes, widgets, and read state.
	Interactions port.InteractionRepository

	// Operations runs search, repair, and counter flushes.
	Operations port.OperationsRepository

	// Assets resolves attachment assets.
	Assets port.AssetResolver

	// Authorizer checks forum permissions.
	Authorizer port.VisibilityAuthorizer

	// Permissions manages forum permission configuration.
	Permissions port.PermissionAdmin

	// Cache caches visible trees.
	Cache port.ReadCache

	// Transactions runs transactional use cases.
	Transactions transaction.Runner

	// Events publishes forum lifecycle events.
	Events emitter.Publisher
}

// NewService creates a forum service facade.
func NewService(deps Dependencies) Service {
	permissions := permissionAdmin(deps)
	return Service{
		Structure: Structure{Service: structureapp.NewService(structureDeps(deps))},
		Content:   Content{Service: contentapp.NewService(contentDeps(deps))},
		Interaction: Interaction{
			Service: interactionapp.NewService(interactionDeps(deps)),
		},
		Operations: Operations{Service: operationsapp.NewService(operationsDeps(deps))},
		Admin: Admin{
			Service: adminapp.NewService(adminDeps(deps, permissions)),
		},
	}
}

// permissionAdmin resolves explicit or authorizer-backed permission administration.
func permissionAdmin(deps Dependencies) port.PermissionAdmin {
	if deps.Permissions != nil {
		return deps.Permissions
	}
	admin, _ := deps.Authorizer.(port.PermissionAdmin)
	return admin
}

// structureDeps adapts facade dependencies to the structure service.
func structureDeps(deps Dependencies) structureapp.Dependencies {
	return structureapp.Dependencies{
		Categories:   deps.Categories,
		Forums:       deps.Forums,
		Authorizer:   deps.Authorizer,
		Cache:        deps.Cache,
		Transactions: deps.Transactions,
		Events:       deps.Events,
	}
}

// contentDeps adapts facade dependencies to the content service.
func contentDeps(deps Dependencies) contentapp.Dependencies {
	return contentapp.Dependencies{
		Forums:       deps.Forums,
		Threads:      deps.Threads,
		Posts:        deps.Posts,
		Assets:       deps.Assets,
		Authorizer:   deps.Authorizer,
		Cache:        deps.Cache,
		Transactions: deps.Transactions,
		Events:       deps.Events,
	}
}

// interactionDeps adapts facade dependencies to the interaction service.
func interactionDeps(deps Dependencies) interactionapp.Dependencies {
	return interactionapp.Dependencies{
		Forums:       deps.Forums,
		Threads:      deps.Threads,
		Posts:        deps.Posts,
		Interactions: deps.Interactions,
		Authorizer:   deps.Authorizer,
		Cache:        deps.Cache,
		Events:       deps.Events,
	}
}

// operationsDeps adapts facade dependencies to the operations service.
func operationsDeps(deps Dependencies) operationsapp.Dependencies {
	return operationsapp.Dependencies{
		Forums:     deps.Forums,
		Operations: deps.Operations,
		Authorizer: deps.Authorizer,
		Cache:      deps.Cache,
		Events:     deps.Events,
	}
}

// adminDeps adapts facade dependencies to the admin service.
func adminDeps(deps Dependencies, permissions port.PermissionAdmin) adminapp.Dependencies {
	return adminapp.Dependencies{
		Forums:       deps.Forums,
		Authorizer:   deps.Authorizer,
		Permissions:  permissions,
		Cache:        deps.Cache,
		Transactions: deps.Transactions,
		Events:       deps.Events,
	}
}

// Ensure Service implements port.Service.
var _ port.Service = Service{}
