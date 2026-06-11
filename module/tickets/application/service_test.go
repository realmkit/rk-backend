package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestCreateTicketValidatesAppealOwnership verifies appeal intake checks punishment ownership.
func TestCreateTicketValidatesAppealOwnership(t *testing.T) {
	service, fakes := newTestService()
	definition := appealDefinition()
	fakes.definitions.items[definition.ID] = definition
	submitter := uuid.New()
	punishmentID := uuid.New()
	assetID := uuid.New()
	fakes.assets.exists[assetID] = true
	fakes.punishments.summary = port.PunishmentSummary{
		ID:           punishmentID,
		TargetUserID: uuid.New(),
	}
	_, err := service.CreateTicket(context.Background(), port.CreateTicketCommand{
		ActorUserID:         submitter,
		DefinitionID:        definition.ID,
		Title:               "Appeal",
		SubmitterUserID:     &submitter,
		PunishmentID:        &punishmentID,
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "please review",
		EvidenceAssetIDs:    []uuid.UUID{assetID},
		IdempotencyKey:      "appeal-1",
	})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("CreateTicket() error = %v, want forbidden", err)
	}
}

// TestCreateTicketCreatesOpenerEvidenceAndIsIdempotent verifies core intake behavior.
func TestCreateTicketCreatesOpenerEvidenceAndIsIdempotent(t *testing.T) {
	service, fakes := newTestService()
	definition := appealDefinition()
	fakes.definitions.items[definition.ID] = definition
	submitter := uuid.New()
	assetID := uuid.New()
	punishmentID := uuid.New()
	fakes.assets.exists[assetID] = true
	fakes.punishments.summary = port.PunishmentSummary{
		ID:           punishmentID,
		TargetUserID: submitter,
	}
	command := port.CreateTicketCommand{
		ActorUserID:         submitter,
		DefinitionID:        definition.ID,
		Title:               "Appeal",
		SubmitterUserID:     &submitter,
		PunishmentID:        &punishmentID,
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "please review",
		EvidenceAssetIDs:    []uuid.UUID{assetID},
		IdempotencyKey:      "appeal-2",
	}
	ticket, err := service.CreateTicket(context.Background(), command)
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	replayed, err := service.CreateTicket(context.Background(), command)
	if err != nil {
		t.Fatalf("CreateTicket() replay error = %v", err)
	}
	if replayed.ID != ticket.ID {
		t.Fatalf("replay ID = %s, want %s", replayed.ID, ticket.ID)
	}
	if got := len(fakes.tickets.messages[ticket.ID]); got != 1 {
		t.Fatalf("messages = %d, want 1", got)
	}
	if got := len(fakes.tickets.evidence[ticket.ID]); got != 1 {
		t.Fatalf("evidence = %d, want 1", got)
	}
}

// TestAcceptAppealRevokesPunishmentAndClearsCache verifies accepted appeal effects.
func TestAcceptAppealRevokesPunishmentAndClearsCache(t *testing.T) {
	service, fakes := newTestService()
	fakes.authorizer.revokePunishment = true
	punishmentID := uuid.New()
	ticket := domain.Ticket{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		Title:        "Appeal",
		Kind:         domain.KindAppeal,
		Status:       domain.StatusOpen,
		PunishmentID: &punishmentID,
		OpenedAt:     time.Now().UTC(),
		Version:      3,
	}.Normalize()
	fakes.tickets.items[ticket.ID] = ticket
	actor := uuid.New()
	updated, err := service.AcceptAppeal(context.Background(), port.AppealDecisionCommand{
		ActorUserID:      actor,
		TicketID:         ticket.ID,
		Reason:           "accepted",
		RevokePunishment: true,
		ExpectedVersion:  3,
		IdempotencyKey:   "accept-1",
	})
	if err != nil {
		t.Fatalf("AcceptAppeal() error = %v", err)
	}
	if updated.Status != domain.StatusAccepted {
		t.Fatalf("Status = %s, want accepted", updated.Status)
	}
	if fakes.punishments.revoked != punishmentID {
		t.Fatalf("revoked = %s, want %s", fakes.punishments.revoked, punishmentID)
	}
	if fakes.cache.ticketClears != 1 || fakes.cache.queueClears != 1 {
		t.Fatalf("cache clears = ticket:%d queue:%d", fakes.cache.ticketClears, fakes.cache.queueClears)
	}
}

// TestAcceptAppealRequiresPunishmentRevocationPermission verifies cross-module appeal effects are separately authorized.
func TestAcceptAppealRequiresPunishmentRevocationPermission(t *testing.T) {
	service, fakes := newTestService()
	fakes.authorizer.revokePunishment = false
	punishmentID := uuid.New()
	ticket := domain.Ticket{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		Title:        "Appeal",
		Kind:         domain.KindAppeal,
		Status:       domain.StatusOpen,
		PunishmentID: &punishmentID,
		OpenedAt:     time.Now().UTC(),
		Version:      3,
	}.Normalize()
	fakes.tickets.items[ticket.ID] = ticket
	_, err := service.AcceptAppeal(context.Background(), port.AppealDecisionCommand{
		ActorUserID:      uuid.New(),
		TicketID:         ticket.ID,
		Reason:           "accepted",
		RevokePunishment: true,
		ExpectedVersion:  3,
		IdempotencyKey:   "accept-denied",
	})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("AcceptAppeal() error = %v, want forbidden", err)
	}
	if fakes.punishments.revoked != uuid.Nil {
		t.Fatalf("revoked = %s, want no punishment revocation", fakes.punishments.revoked)
	}
	if fakes.tickets.items[ticket.ID].Status != domain.StatusOpen {
		t.Fatalf("ticket status = %s, want unchanged open", fakes.tickets.items[ticket.ID].Status)
	}
}

// TestTicketServiceFailsClosedWithoutAuthorizer verifies private ticket workflows require an authorizer dependency.
func TestTicketServiceFailsClosedWithoutAuthorizer(t *testing.T) {
	service, fakes := newTestService()
	service.authorizer = nil
	definition := appealDefinition()
	definition.RequiresEvidence = false
	fakes.definitions.items[definition.ID] = definition
	submitter := uuid.New()
	punishmentID := uuid.New()
	fakes.punishments.summary = port.PunishmentSummary{
		ID:           punishmentID,
		TargetUserID: submitter,
	}
	_, err := service.CreateTicket(context.Background(), port.CreateTicketCommand{
		ActorUserID:         submitter,
		DefinitionID:        definition.ID,
		Title:               "Appeal",
		SubmitterUserID:     &submitter,
		PunishmentID:        &punishmentID,
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "please review",
		IdempotencyKey:      "appeal-no-authz",
	})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("CreateTicket() error = %v, want forbidden", err)
	}
	_, err = service.GetTicket(context.Background(), uuid.New(), submitter)
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("GetTicket() error = %v, want forbidden", err)
	}
	_, err = service.CreateMessage(context.Background(), port.MessageCommand{
		ActorUserID:         submitter,
		TicketID:            uuid.New(),
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "hello",
	})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("CreateMessage() error = %v, want forbidden", err)
	}
}

// TestCreateMessageAndEvidenceValidatePermissionsAndAssets verifies conversation use cases.
func TestCreateMessageAndEvidenceValidatePermissionsAndAssets(t *testing.T) {
	service, fakes := newTestService()
	ticketID := uuid.New()
	actor := uuid.New()
	fakes.authorizer.reply = true
	message, err := service.CreateMessage(context.Background(), port.MessageCommand{
		ActorUserID:         actor,
		TicketID:            ticketID,
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         "hello",
		IdempotencyKey:      "msg-1",
	})
	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}
	assetID := uuid.New()
	fakes.assets.exists[assetID] = true
	if _, err := service.AddEvidence(context.Background(), port.EvidenceCommand{
		ActorUserID: actor,
		TicketID:    ticketID,
		MessageID:   &message.ID,
		AssetID:     &assetID,
	}); err != nil {
		t.Fatalf("AddEvidence() error = %v", err)
	}
	if fakes.cache.ticketClears != 2 {
		t.Fatalf("ticket cache clears = %d, want 2", fakes.cache.ticketClears)
	}
}

// TestOperationsDelegateToRepositories verifies operational use cases.
func TestOperationsDelegateToRepositories(t *testing.T) {
	service, fakes := newTestService()
	fakes.tickets.report = domain.DriftReport{Mismatches: []string{"drift"}}
	report, err := service.RebuildStats(context.Background())
	if err != nil {
		t.Fatalf("RebuildStats() error = %v", err)
	}
	if !report.Repaired || len(report.Mismatches) != 1 {
		t.Fatalf("report = %+v, want repaired drift", report)
	}
	if fakes.cache.allClears != 1 {
		t.Fatalf("all cache clears = %d, want 1", fakes.cache.allClears)
	}
}

// TestDefinitionLifecycle verifies definition use cases.
func TestDefinitionLifecycle(t *testing.T) {
	service, _ := newTestService()
	definition := domain.Definition{
		ID:      uuid.New(),
		Key:     "support",
		Name:    "Support",
		Kind:    domain.KindSupport,
		Status:  domain.DefinitionActive,
		Version: 1,
	}
	created, err := service.CreateDefinition(context.Background(), definition)
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	created.Name = "Help"
	updated, err := service.UpdateDefinition(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("UpdateDefinition() error = %v", err)
	}
	if updated.Name != "Help" {
		t.Fatalf("Name = %q, want Help", updated.Name)
	}
	if _, err := service.GetDefinition(context.Background(), created.ID); err != nil {
		t.Fatalf("GetDefinition() error = %v", err)
	}
	if list, err := service.ListDefinitions(context.Background(), port.DefinitionFilter{}, pagination.Page{Limit: 10}); err != nil ||
		len(list.Items) != 1 {
		t.Fatalf("ListDefinitions() = %+v, %v; want one item", list, err)
	}
	if err := service.DeleteDefinition(context.Background(), created.ID, updated.Version); err != nil {
		t.Fatalf("DeleteDefinition() error = %v", err)
	}
}

// TestStaffWorkflowTransitions verifies assignment, escalation, close, reopen and reject.
func TestStaffWorkflowTransitions(t *testing.T) {
	service, fakes := newTestService()
	ticket := domain.Ticket{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		Title:        "Support",
		Kind:         domain.KindSupport,
		Status:       domain.StatusOpen,
		OpenedAt:     time.Now().UTC(),
		Version:      1,
	}.Normalize()
	fakes.tickets.items[ticket.ID] = ticket
	actor := uuid.New()
	assignee := uuid.New()
	team := uuid.New()
	ticket, err := service.AssignTicket(context.Background(), port.StaffCommand{
		ActorUserID:     actor,
		TicketID:        ticket.ID,
		AssigneeUserID:  &assignee,
		ExpectedVersion: ticket.Version,
	})
	if err != nil {
		t.Fatalf("AssignTicket() error = %v", err)
	}
	ticket, err = service.EscalateTicket(context.Background(), port.StaffCommand{
		ActorUserID:     actor,
		TicketID:        ticket.ID,
		TeamGroupID:     &team,
		ExpectedVersion: ticket.Version,
	})
	if err != nil {
		t.Fatalf("EscalateTicket() error = %v", err)
	}
	ticket, err = service.CloseTicket(context.Background(), port.StaffCommand{
		ActorUserID:     actor,
		TicketID:        ticket.ID,
		Reason:          "done",
		ExpectedVersion: ticket.Version,
	})
	if err != nil {
		t.Fatalf("CloseTicket() error = %v", err)
	}
	ticket, err = service.ReopenTicket(context.Background(), port.StaffCommand{
		ActorUserID:     actor,
		TicketID:        ticket.ID,
		ExpectedVersion: ticket.Version,
	})
	if err != nil {
		t.Fatalf("ReopenTicket() error = %v", err)
	}
	ticket.Kind = domain.KindAppeal
	fakes.tickets.items[ticket.ID] = ticket
	rejected, err := service.RejectAppeal(context.Background(), port.AppealDecisionCommand{
		ActorUserID:     actor,
		TicketID:        ticket.ID,
		Reason:          "no",
		ExpectedVersion: ticket.Version,
	})
	if err != nil {
		t.Fatalf("RejectAppeal() error = %v", err)
	}
	if rejected.Status != domain.StatusRejected || rejected.Resolution != "no" {
		t.Fatalf("rejected = %+v, want rejected appeal", rejected)
	}
}

// testFakes groups application fake dependencies.
type testFakes struct {
	definitions *definitionRepo
	tickets     *ticketRepo
	cache       *cacheFake
	assets      *assetFake
	punishments *punishmentFake
	authorizer  *authorizerFake
}

// newTestService creates a service with in-memory ports.
func newTestService() (Service, testFakes) {
	fakes := testFakes{
		definitions: &definitionRepo{items: map[uuid.UUID]domain.Definition{}},
		tickets:     newTicketRepo(),
		cache:       &cacheFake{},
		assets:      &assetFake{exists: map[uuid.UUID]bool{}},
		punishments: &punishmentFake{},
		authorizer: &authorizerFake{
			create:           true,
			view:             true,
			reply:            true,
			staff:            true,
			revokePunishment: true,
		},
	}
	service := NewService(Dependencies{
		Definitions: fakes.definitions,
		Tickets:     fakes.tickets,
		Cache:       fakes.cache,
		Assets:      fakes.assets,
		Punishments: fakes.punishments,
		Authorizer:  fakes.authorizer,
	})
	return service, fakes
}

// appealDefinition returns a valid appeal definition.
func appealDefinition() domain.Definition {
	return domain.Definition{
		ID:                 uuid.New(),
		Key:                "appeal",
		Name:               "Appeal",
		Kind:               domain.KindAppeal,
		Status:             domain.DefinitionActive,
		RequiresPunishment: true,
		RequiresEvidence:   true,
		Version:            1,
	}
}
