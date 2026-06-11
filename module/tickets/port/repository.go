package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// DefinitionRepository stores ticket definitions.
type DefinitionRepository interface {
	Create(context.Context, domain.Definition) (domain.Definition, error)
	Update(context.Context, domain.Definition, uint64) (domain.Definition, error)
	Delete(context.Context, uuid.UUID, uint64) error
	FindByID(context.Context, uuid.UUID) (domain.Definition, error)
	List(context.Context, DefinitionFilter, pagination.Page) (pagination.Result[domain.Definition], error)
}

// TicketRepository stores ticket cases and read models.
type TicketRepository interface {
	Create(context.Context, domain.Ticket, domain.Message, []domain.Evidence) (domain.Ticket, error)
	Update(context.Context, domain.Ticket, uint64) (domain.Ticket, error)
	FindByID(context.Context, uuid.UUID) (domain.Ticket, error)
	FindByIdempotencyKey(context.Context, string) (domain.Ticket, error)
	List(context.Context, TicketFilter, pagination.Page) (pagination.Result[domain.Ticket], error)
	AddMessage(context.Context, domain.Message) (domain.Message, error)
	ListMessages(context.Context, uuid.UUID, bool, pagination.Page) (pagination.Result[domain.Message], error)
	AddEvidence(context.Context, domain.Evidence) (domain.Evidence, error)
	ListEvidence(context.Context, uuid.UUID, bool) ([]domain.Evidence, error)
	AddAction(context.Context, domain.Action) (domain.Action, error)
	ListActions(context.Context, uuid.UUID, pagination.Page) (pagination.Result[domain.Action], error)
	VerifyStats(context.Context) (domain.DriftReport, error)
	RebuildStats(context.Context) (domain.DriftReport, error)
	DetectSLABreaches(context.Context, time.Time) ([]domain.Ticket, error)
	CloseStale(context.Context, time.Time) (int64, error)
}

// Cache caches ticket read models.
type Cache interface {
	ClearTicket(context.Context, uuid.UUID) error
	ClearQueues(context.Context) error
	ClearAll(context.Context) error
}

// TransactionRunner runs work in a transaction.
type TransactionRunner interface {
	WithinTx(context.Context, func(context.Context) error) error
}
