package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// definitionRepo is an in-memory definition repository.
type definitionRepo struct {
	items map[uuid.UUID]domain.Definition
}

// Create stores one definition.
func (repo *definitionRepo) Create(_ context.Context, definition domain.Definition) (domain.Definition, error) {
	repo.items[definition.ID] = definition
	return definition, nil
}

// Update stores one definition.
func (repo *definitionRepo) Update(_ context.Context, definition domain.Definition, _ uint64) (domain.Definition, error) {
	definition.Version++
	repo.items[definition.ID] = definition
	return definition, nil
}

// Delete deletes one definition.
func (repo *definitionRepo) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repo.items, id)
	return nil
}

// FindByID returns one definition.
func (repo *definitionRepo) FindByID(_ context.Context, id uuid.UUID) (domain.Definition, error) {
	item, ok := repo.items[id]
	if !ok {
		return domain.Definition{}, port.ErrNotFound
	}
	return item, nil
}

// List returns all definitions.
func (repo *definitionRepo) List(_ context.Context, _ port.DefinitionFilter, _ pagination.Page) (pagination.Result[domain.Definition], error) {
	items := make([]domain.Definition, 0, len(repo.items))
	for _, item := range repo.items {
		items = append(items, item)
	}
	return pagination.Result[domain.Definition]{Items: items}, nil
}

// ticketRepo is an in-memory ticket repository.
type ticketRepo struct {
	items    map[uuid.UUID]domain.Ticket
	messages map[uuid.UUID][]domain.Message
	evidence map[uuid.UUID][]domain.Evidence
	actions  map[uuid.UUID][]domain.Action
	report   domain.DriftReport
}

// newTicketRepo creates a ticket fake.
func newTicketRepo() *ticketRepo {
	return &ticketRepo{
		items:    map[uuid.UUID]domain.Ticket{},
		messages: map[uuid.UUID][]domain.Message{},
		evidence: map[uuid.UUID][]domain.Evidence{},
		actions:  map[uuid.UUID][]domain.Action{},
	}
}

// Create stores one ticket graph.
func (repo *ticketRepo) Create(_ context.Context, ticket domain.Ticket, opener domain.Message, evidence []domain.Evidence) (domain.Ticket, error) {
	opener.Sequence = 1
	repo.items[ticket.ID] = ticket
	repo.messages[ticket.ID] = append(repo.messages[ticket.ID], opener)
	repo.evidence[ticket.ID] = append(repo.evidence[ticket.ID], evidence...)
	return ticket, nil
}

// Update stores one ticket.
func (repo *ticketRepo) Update(_ context.Context, ticket domain.Ticket, expected uint64) (domain.Ticket, error) {
	if repo.items[ticket.ID].Version != expected {
		return domain.Ticket{}, port.ErrPreconditionFailed
	}
	ticket.Version = expected + 1
	repo.items[ticket.ID] = ticket
	return ticket, nil
}

// FindByID returns one ticket.
func (repo *ticketRepo) FindByID(_ context.Context, id uuid.UUID) (domain.Ticket, error) {
	item, ok := repo.items[id]
	if !ok {
		return domain.Ticket{}, port.ErrNotFound
	}
	return item, nil
}

// FindByIdempotencyKey returns one ticket by key.
func (repo *ticketRepo) FindByIdempotencyKey(_ context.Context, key string) (domain.Ticket, error) {
	for _, item := range repo.items {
		if item.Key == key {
			return item, nil
		}
	}
	return domain.Ticket{}, port.ErrNotFound
}

// List returns all tickets.
func (repo *ticketRepo) List(_ context.Context, _ port.TicketFilter, _ pagination.Page) (pagination.Result[domain.Ticket], error) {
	items := make([]domain.Ticket, 0, len(repo.items))
	for _, item := range repo.items {
		items = append(items, item)
	}
	return pagination.Result[domain.Ticket]{Items: items}, nil
}

// AddMessage stores one message.
func (repo *ticketRepo) AddMessage(_ context.Context, message domain.Message) (domain.Message, error) {
	message.Sequence = int64(len(repo.messages[message.TicketID]) + 1)
	repo.messages[message.TicketID] = append(repo.messages[message.TicketID], message)
	return message, nil
}

// ListMessages returns ticket messages.
func (repo *ticketRepo) ListMessages(_ context.Context, id uuid.UUID, _ bool, _ pagination.Page) (pagination.Result[domain.Message], error) {
	return pagination.Result[domain.Message]{Items: repo.messages[id]}, nil
}

// AddEvidence stores one evidence record.
func (repo *ticketRepo) AddEvidence(_ context.Context, evidence domain.Evidence) (domain.Evidence, error) {
	repo.evidence[evidence.TicketID] = append(repo.evidence[evidence.TicketID], evidence)
	return evidence, nil
}

// ListEvidence returns evidence records.
func (repo *ticketRepo) ListEvidence(_ context.Context, id uuid.UUID, _ bool) ([]domain.Evidence, error) {
	return repo.evidence[id], nil
}

// AddAction stores one action.
func (repo *ticketRepo) AddAction(_ context.Context, action domain.Action) (domain.Action, error) {
	repo.actions[action.TicketID] = append(repo.actions[action.TicketID], action)
	return action, nil
}

// ListActions returns ticket actions.
func (repo *ticketRepo) ListActions(_ context.Context, id uuid.UUID, _ pagination.Page) (pagination.Result[domain.Action], error) {
	return pagination.Result[domain.Action]{Items: repo.actions[id]}, nil
}

// VerifyStats returns a configured drift report.
func (repo *ticketRepo) VerifyStats(context.Context) (domain.DriftReport, error) {
	return repo.report, nil
}

// RebuildStats returns a repaired configured drift report.
func (repo *ticketRepo) RebuildStats(context.Context) (domain.DriftReport, error) {
	repo.report.Repaired = true
	return repo.report, nil
}

// DetectSLABreaches returns no fake breaches.
func (repo *ticketRepo) DetectSLABreaches(context.Context, time.Time) ([]domain.Ticket, error) {
	return nil, nil
}

// CloseStale returns no fake changes.
func (repo *ticketRepo) CloseStale(context.Context, time.Time) (int64, error) {
	return 0, nil
}

// cacheFake records cache invalidations.
type cacheFake struct {
	ticketClears int
	queueClears  int
	allClears    int
}

// ClearTicket records a ticket cache clear.
func (cache *cacheFake) ClearTicket(context.Context, uuid.UUID) error {
	cache.ticketClears++
	return nil
}

// ClearQueues records a queue cache clear.
func (cache *cacheFake) ClearQueues(context.Context) error {
	cache.queueClears++
	return nil
}

// ClearAll records a global cache clear.
func (cache *cacheFake) ClearAll(context.Context) error {
	cache.allClears++
	return nil
}

// assetFake checks asset existence.
type assetFake struct{ exists map[uuid.UUID]bool }

// AssetExists reports whether an asset exists.
func (fake *assetFake) AssetExists(_ context.Context, id uuid.UUID) (bool, error) {
	return fake.exists[id], nil
}

// punishmentFake records punishment operations.
type punishmentFake struct {
	summary port.PunishmentSummary
	revoked uuid.UUID
}

// GetPunishment returns one configured punishment.
func (fake *punishmentFake) GetPunishment(context.Context, uuid.UUID) (port.PunishmentSummary, error) {
	return fake.summary, nil
}

// RevokePunishment records a revocation.
func (fake *punishmentFake) RevokePunishment(_ context.Context, id uuid.UUID, _ uuid.UUID, _ string, _ uint64) error {
	fake.revoked = id
	return nil
}

// authorizerFake returns configured permission decisions.
type authorizerFake struct {
	create bool
	view   bool
	reply  bool
	staff  bool
}

// CanCreate reports fake create access.
func (fake *authorizerFake) CanCreate(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return fake.create, nil
}

// CanView reports fake view access.
func (fake *authorizerFake) CanView(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return fake.view, nil
}

// CanReply reports fake reply access.
func (fake *authorizerFake) CanReply(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return fake.reply, nil
}

// CanStaffAction reports fake staff action access.
func (fake *authorizerFake) CanStaffAction(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return fake.staff, nil
}
