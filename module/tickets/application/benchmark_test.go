package application

import (
	"encoding/json"
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
