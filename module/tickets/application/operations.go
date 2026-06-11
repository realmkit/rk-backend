package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
)

// VerifyStats reports ticket counter drift.
func (service Service) VerifyStats(ctx context.Context) (domain.DriftReport, error) {
	return service.tickets.VerifyStats(ctx)
}

// RebuildStats repairs ticket counter drift.
func (service Service) RebuildStats(ctx context.Context) (domain.DriftReport, error) {
	report, err := service.tickets.RebuildStats(ctx)
	if err != nil {
		return domain.DriftReport{}, err
	}
	_ = service.ClearCache(ctx)
	return report, service.publishOps(ctx, "tickets.stats.rebuilt", int64(len(report.Mismatches)))
}

// DetectSLABreaches detects overdue tickets and emits events.
func (service Service) DetectSLABreaches(ctx context.Context) (int64, error) {
	tickets, err := service.tickets.DetectSLABreaches(ctx, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	for _, ticket := range tickets {
		_ = service.publishTicket(ctx, "tickets.sla.breached", ticket)
	}
	return int64(len(tickets)), nil
}

// CloseStaleTickets closes stale pending-submitter tickets.
func (service Service) CloseStaleTickets(ctx context.Context) (int64, error) {
	count, err := service.tickets.CloseStale(ctx, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	if count > 0 {
		_ = service.ClearCache(ctx)
	}
	return count, nil
}

// ClearCache clears ticket read caches.
func (service Service) ClearCache(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearAll(ctx)
}

// publishDefinition emits a private definition lifecycle event.
func (service Service) publishDefinition(ctx context.Context, key string, definition domain.Definition) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           eventdomain.EventKey(key),
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerTickets,
		AggregateType: "ticket_definition",
		AggregateID:   emitter.UUID(definition.ID),
		Payload:       definition,
		Scopes:        []eventdomain.Scope{{Type: eventdomain.ScopeSystem}},
		AvailableAt:   time.Now().UTC(),
	})
}

// publishTicket emits a ticket lifecycle event scoped to participants and staff.
func (service Service) publishTicket(ctx context.Context, key string, ticket domain.Ticket) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           eventdomain.EventKey(key),
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerTickets,
		AggregateType: "ticket",
		AggregateID:   emitter.UUID(ticket.ID),
		Payload:       ticket,
		Scopes:        ticketScopes(ticket),
		AvailableAt:   time.Now().UTC(),
	})
}

// publishMessage emits a ticket message event scoped to the ticket.
func (service Service) publishMessage(ctx context.Context, key string, message domain.Message) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           eventdomain.EventKey(key),
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerTickets,
		AggregateType: "ticket_message",
		AggregateID:   emitter.UUID(message.ID),
		Payload:       message,
		Scopes:        []eventdomain.Scope{ticketScope(message.TicketID)},
		AvailableAt:   time.Now().UTC(),
	})
}

// publishEvidence emits an evidence event scoped to the ticket.
func (service Service) publishEvidence(ctx context.Context, key string, evidence domain.Evidence) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           eventdomain.EventKey(key),
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerTickets,
		AggregateType: "ticket_evidence",
		AggregateID:   emitter.UUID(evidence.ID),
		Payload:       evidence,
		Scopes:        []eventdomain.Scope{ticketScope(evidence.TicketID)},
		AvailableAt:   time.Now().UTC(),
	})
}

// publishOps emits a system-only operational ticket event.
func (service Service) publishOps(ctx context.Context, key string, count int64) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           eventdomain.EventKey(key),
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerTickets,
		AggregateType: "ticket_stats",
		Payload:       map[string]int64{"count": count},
		Scopes:        []eventdomain.Scope{{Type: eventdomain.ScopeSystem}},
		AvailableAt:   time.Now().UTC(),
	})
}

// ticketScopes returns the websocket/event audiences for a ticket.
func ticketScopes(ticket domain.Ticket) []eventdomain.Scope {
	scopes := []eventdomain.Scope{
		ticketScope(ticket.ID),
		{Type: eventdomain.ScopeStaff},
	}
	if ticket.SubmitterUserID != nil {
		scopes = append(scopes, eventdomain.Scope{
			Type: eventdomain.ScopeUser,
			ID:   ticket.SubmitterUserID.String(),
		})
	}
	if ticket.AssigneeUserID != nil {
		scopes = append(scopes, eventdomain.Scope{
			Type: eventdomain.ScopeUser,
			ID:   ticket.AssigneeUserID.String(),
		})
	}
	if ticket.CurrentTeamGroupID != nil {
		scopes = append(scopes, eventdomain.Scope{
			Type: eventdomain.ScopeGroup,
			ID:   ticket.CurrentTeamGroupID.String(),
		})
	}
	return scopes
}

// ticketScope returns the private ticket aggregate scope.
func ticketScope(ticketID uuid.UUID) eventdomain.Scope {
	return eventdomain.Scope{Type: eventdomain.ScopeTicket, ID: ticketID.String()}
}
