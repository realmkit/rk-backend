package application

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
)

// benchmarkTicket stores the prepared ticket benchmark result.
var benchmarkTicket domain.Ticket

// benchmarkMessage stores the prepared opener message benchmark result.
var benchmarkMessage domain.Message

// benchmarkEvidence stores the prepared ticket evidence benchmark result.
var benchmarkEvidence []domain.Evidence

// BenchmarkPrepareTicket measures ticket, opener message, and evidence preparation.
func BenchmarkPrepareTicket(b *testing.B) {
	service := Service{}
	definition := appealDefinition()
	actorID := uuid.New()
	submitterID := uuid.New()
	targetID := uuid.New()
	punishmentID := uuid.New()
	command := port.CreateTicketCommand{
		ActorUserID:         actorID,
		DefinitionID:        definition.ID,
		Title:               "Appeal benchmark",
		SubmitterUserID:     &submitterID,
		TargetUserID:        &targetID,
		PunishmentID:        &punishmentID,
		ContentDocumentJSON: json.RawMessage(`{"type":"doc","content":[{"type":"text","text":"hello"}]}`),
		ContentText:         "hello",
		EvidenceAssetIDs:    []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
		IdempotencyKey:      "benchmark",
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		ticket, message, evidence := service.prepareTicket(command, definition)
		benchmarkTicket = ticket
		benchmarkMessage = message
		benchmarkEvidence = evidence
	}
}

// BenchmarkCreateTicket measures full ticket intake orchestration over in-memory ports.
func BenchmarkCreateTicket(b *testing.B) {
	service, fakes := newTestService()
	definition := appealDefinition()
	fakes.definitions.items[definition.ID] = definition
	submitterID := uuid.New()
	punishmentID := uuid.New()
	assetID := uuid.New()
	fakes.assets.exists[assetID] = true
	fakes.punishments.summary = port.PunishmentSummary{
		ID:           punishmentID,
		TargetUserID: submitterID,
	}
	command := port.CreateTicketCommand{
		ActorUserID:         submitterID,
		DefinitionID:        definition.ID,
		Title:               "Appeal benchmark",
		SubmitterUserID:     &submitterID,
		PunishmentID:        &punishmentID,
		ContentDocumentJSON: json.RawMessage(`{"type":"doc","content":[{"type":"text","text":"hello"}]}`),
		ContentText:         "hello",
		EvidenceAssetIDs:    []uuid.UUID{assetID},
	}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		command.IdempotencyKey = "benchmark-" + strconv.Itoa(index)
		ticket, err := service.CreateTicket(ctx, command)
		if err != nil {
			b.Fatalf("CreateTicket() error = %v", err)
		}
		delete(fakes.tickets.items, ticket.ID)
		delete(fakes.tickets.messages, ticket.ID)
		delete(fakes.tickets.evidence, ticket.ID)
		benchmarkTicket = ticket
	}
}
