// Package application coordinates ticket use cases.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
)

// Dependencies contains ticket service dependencies.
type Dependencies struct {
	Definitions  port.DefinitionRepository
	Tickets      port.TicketRepository
	Cache        port.Cache
	Transactions port.TransactionRunner
	Events       emitter.Publisher
	Authorizer   port.Authorizer
	Assets       port.AssetResolver
	Users        port.UserResolver
	Groups       port.GroupResolver
	Punishments  interface {
		port.PunishmentReader
		port.PunishmentExecutor
	}
}

// Service implements ticket use cases.
type Service struct {
	definitions  port.DefinitionRepository
	tickets      port.TicketRepository
	cache        port.Cache
	transactions port.TransactionRunner
	events       emitter.Publisher
	authorizer   port.Authorizer
	assets       port.AssetResolver
	users        port.UserResolver
	groups       port.GroupResolver
	punishments  interface {
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
