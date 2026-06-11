package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// DefinitionService exposes ticket definition use cases.
type DefinitionService interface {
	CreateDefinition(context.Context, domain.Definition) (domain.Definition, error)
	UpdateDefinition(context.Context, domain.Definition, uint64) (domain.Definition, error)
	DeleteDefinition(context.Context, uuid.UUID, uint64) error
	GetDefinition(context.Context, uuid.UUID) (domain.Definition, error)
	ListDefinitions(context.Context, DefinitionFilter, pagination.Page) (pagination.Result[domain.Definition], error)
}

// TicketService exposes ticket intake and staff use cases.
type TicketService interface {
	CreateTicket(context.Context, CreateTicketCommand) (domain.Ticket, error)
	GetTicket(context.Context, uuid.UUID, uuid.UUID) (domain.Ticket, error)
	ListTickets(context.Context, TicketFilter, pagination.Page) (pagination.Result[domain.Ticket], error)
	AssignTicket(context.Context, StaffCommand) (domain.Ticket, error)
	EscalateTicket(context.Context, StaffCommand) (domain.Ticket, error)
	CloseTicket(context.Context, StaffCommand) (domain.Ticket, error)
	ReopenTicket(context.Context, StaffCommand) (domain.Ticket, error)
	AcceptAppeal(context.Context, AppealDecisionCommand) (domain.Ticket, error)
	RejectAppeal(context.Context, AppealDecisionCommand) (domain.Ticket, error)
}

// ConversationService exposes ticket conversation use cases.
type ConversationService interface {
	CreateMessage(context.Context, MessageCommand) (domain.Message, error)
	ListMessages(context.Context, uuid.UUID, uuid.UUID, bool, pagination.Page) (pagination.Result[domain.Message], error)
	AddEvidence(context.Context, EvidenceCommand) (domain.Evidence, error)
	ListEvidence(context.Context, uuid.UUID, uuid.UUID, bool) ([]domain.Evidence, error)
}

// OperationsService exposes operational ticket use cases.
type OperationsService interface {
	VerifyStats(context.Context) (domain.DriftReport, error)
	RebuildStats(context.Context) (domain.DriftReport, error)
	DetectSLABreaches(context.Context) (int64, error)
	CloseStaleTickets(context.Context) (int64, error)
	ClearCache(context.Context) error
}
