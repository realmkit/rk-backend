// Package application coordinates ticket use cases.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// Dependencies contains ticket service dependencies.
type Dependencies struct {
	Definitions  port.DefinitionRepository // Definitions stores the definitions value.
	Tickets      port.TicketRepository     // Tickets stores the tickets value.
	Cache        port.Cache                // Cache stores the cache value.
	Transactions port.TransactionRunner    // Transactions stores the transactions value.
	Events       emitter.Publisher         // Events stores the events value.
	Authorizer   port.Authorizer           // Authorizer stores the authorizer value.
	Assets       port.AssetResolver        // Assets stores the assets value.
	Users        port.UserResolver         // Users stores the users value.
	Groups       port.GroupResolver        // Groups stores the groups value.
	// Punishments resolves and mutates linked punishments.
	Punishments interface {
		port.PunishmentReader
		port.PunishmentExecutor
	}
}

// Service implements ticket use cases.
type Service struct {
	definitions  port.DefinitionRepository // definitions stores the definitions value.
	tickets      port.TicketRepository     // tickets stores the tickets value.
	cache        port.Cache                // cache stores the cache value.
	transactions port.TransactionRunner    // transactions stores the transactions value.
	events       emitter.Publisher         // events stores the events value.
	authorizer   port.Authorizer           // authorizer stores the authorizer value.
	assets       port.AssetResolver        // assets stores the assets value.
	users        port.UserResolver         // users stores the users value.
	groups       port.GroupResolver        // groups stores the groups value.
	// punishments resolves and mutates linked punishments.
	punishments interface {
		port.PunishmentReader
		port.PunishmentExecutor
	}
}

// NewService creates a ticket service.
func NewService(deps Dependencies) Service {
	return Service{
		definitions:  deps.Definitions,
		tickets:      deps.Tickets,
		cache:        deps.Cache,
		transactions: deps.Transactions,
		events:       deps.Events,
		authorizer:   deps.Authorizer,
		assets:       deps.Assets,
		users:        deps.Users,
		groups:       deps.Groups,
		punishments:  deps.Punishments,
	}
}

// withinTx runs fn in a transaction when configured.
func (service Service) withinTx(ctx context.Context, fn func(context.Context) error) error {
	if service.transactions == nil {
		return fn(ctx)
	}
	return service.transactions.WithinTx(ctx, fn)
}

// clearTicket clears ticket read caches.
func (service Service) clearTicket(ctx context.Context, ticketID uuid.UUID) error {
	if service.cache == nil {
		return nil
	}
	if err := service.cache.ClearTicket(ctx, ticketID); err != nil {
		return err
	}
	return service.cache.ClearQueues(ctx)
}

// requireAuthorizer reports forbidden when a security-sensitive path lacks an authorizer.
func (service Service) requireAuthorizer() error {
	if service.authorizer == nil {
		return port.ErrForbidden
	}
	return nil
}

// can reports true when check allows.
func can(check func() (bool, error)) error {
	allowed, err := check()
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}
