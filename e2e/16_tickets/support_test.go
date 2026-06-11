// Package tickets_e2e verifies ticket and appeal journeys through the real server.
package tickets_e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapplication "github.com/realmkit/rk-backend/module/groups/application"
	punishmentspostgres "github.com/realmkit/rk-backend/module/punishments/adapter/postgres"
	punishmentsredis "github.com/realmkit/rk-backend/module/punishments/adapter/redis"
	punishmentsapplication "github.com/realmkit/rk-backend/module/punishments/application"
	ticketgroups "github.com/realmkit/rk-backend/module/tickets/adapter/groups"
	ticketshttp "github.com/realmkit/rk-backend/module/tickets/adapter/http"
	ticketspostgres "github.com/realmkit/rk-backend/module/tickets/adapter/postgres"
	ticketpunishments "github.com/realmkit/rk-backend/module/tickets/adapter/punishments"
	ticketsredis "github.com/realmkit/rk-backend/module/tickets/adapter/redis"
	ticketsapplication "github.com/realmkit/rk-backend/module/tickets/application"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/transaction"
	goredis "github.com/redis/go-redis/v9"
)

// ticketsFixture contains ticket e2e wiring.
type ticketsFixture struct {
	ecosystem   *harness.Ecosystem
	service     ticketsapplication.Service
	groups      groupsapplication.Service
	punishments punishmentsapplication.Service
	resolver    *ticketResolver
	events      *eventtesting.PublisherRecorder
}

// newTicketsFixture starts a server with tickets and backing services.
func newTicketsFixture(t *testing.T) ticketsFixture {
	t.Helper()
	database := harness.NewSQLiteDatabase(t)
	events := &eventtesting.PublisherRecorder{}
	redisServer := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
		redisServer.Close()
	})
	groups := groupsapplication.NewService(
		groupspostgres.NewGroupRepository(database.Store),
		groupspostgres.NewMembershipRepository(database.Store),
		groupspostgres.NewTupleRepository(database.Store),
		groupspostgres.NewPermissionRepository(database.Store),
	)
	punishments := punishmentsapplication.NewService(punishmentsapplication.Dependencies{
		Definitions:  punishmentspostgres.NewDefinitionRepository(database.Store),
		Cases:        punishmentspostgres.NewCaseRepository(database.Store),
		Cache:        punishmentsredis.NewCache(redisClient),
		Transactions: transaction.New(database.DB),
		Events:       events,
	})
	resolver := &ticketResolver{assets: map[uuid.UUID]bool{}}
	service := ticketsapplication.NewService(ticketsapplication.Dependencies{
		Definitions:  ticketspostgres.NewDefinitionRepository(database.Store),
		Tickets:      ticketspostgres.NewTicketRepository(database.Store),
		Cache:        ticketsredis.NewCache(redisClient),
		Transactions: transaction.New(database.DB),
		Events:       events,
		Authorizer:   ticketgroups.NewAuthorizer(groups),
		Assets:       resolver,
		Users:        resolver,
		Groups:       resolver,
		Punishments:  ticketpunishments.NewResolver(punishments),
	})
	ecosystem := harness.New(
		t,
		harness.WithDatabase(database),
		harness.WithServerOptions(server.WithTickets(ticketshttp.Services{
			Definitions:  service,
			Tickets:      service,
			Conversation: service,
			Operations:   service,
		})),
	)
	return ticketsFixture{
		ecosystem: ecosystem, service: service, groups: groups,
		punishments: punishments, resolver: resolver, events: events,
	}
}

// ticketResolver is an in-memory resolver for external identities/assets.
type ticketResolver struct {
	assets map[uuid.UUID]bool
}

// AssetExists reports whether an evidence asset is known.
func (resolver *ticketResolver) AssetExists(_ context.Context, id uuid.UUID) (bool, error) {
	return resolver.assets[id], nil
}

// UserExists reports test users as resolvable.
func (resolver *ticketResolver) UserExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}

// GroupExists reports test groups as resolvable.
func (resolver *ticketResolver) GroupExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}

// do sends a request through the fixture server.
func (fixture ticketsFixture) do(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	return fixture.ecosystem.Test(t, request)
}
