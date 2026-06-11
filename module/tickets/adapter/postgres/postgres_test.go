package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDefinitionRepositoryLifecycle verifies definition persistence.
func TestDefinitionRepositoryLifecycle(t *testing.T) {
	repo := NewDefinitionRepository(orm.NewStore(newDB(t)))
	definition := validDefinition()
	created, err := repo.Create(context.Background(), definition)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	created.Name = "Updated"
	updated, err := repo.Update(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("Version = %d, want 2", updated.Version)
	}
	result, err := repo.List(context.Background(), port.DefinitionFilter{Kind: domain.KindSupport}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(result.Items))
	}
	if err := repo.Delete(context.Background(), created.ID, updated.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

// TestTicketRepositoryCreateMessageEvidenceAndActions verifies ticket graph persistence.
func TestTicketRepositoryCreateMessageEvidenceAndActions(t *testing.T) {
	db := newDB(t)
	repo := NewTicketRepository(orm.NewStore(db))
	ticket := validTicket()
	opener := validMessage(ticket.ID, "opener")
	assetID := uuid.New()
	evidence := domain.Evidence{
		ID:         uuid.New(),
		TicketID:   ticket.ID,
		AssetID:    &assetID,
		Visibility: domain.VisibilityParticipants,
		CreatedAt:  time.Now().UTC(),
	}
	created, err := repo.Create(context.Background(), ticket, opener, []domain.Evidence{evidence})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != ticket.ID {
		t.Fatalf("created ID = %s, want %s", created.ID, ticket.ID)
	}
	reply, err := repo.AddMessage(context.Background(), validMessage(ticket.ID, "reply"))
	if err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}
	if reply.Sequence != 2 {
		t.Fatalf("reply.Sequence = %d, want 2", reply.Sequence)
	}
	messages, err := repo.ListMessages(context.Background(), ticket.ID, false, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages.Items) != 2 {
		t.Fatalf("messages = %d, want 2", len(messages.Items))
	}
	if _, err := repo.AddAction(context.Background(), domain.Action{
		ID:          uuid.New(),
		TicketID:    ticket.ID,
		Type:        domain.ActionClose,
		Status:      domain.ActionCompleted,
		PayloadJSON: []byte(`{}`),
		CreatedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("AddAction() error = %v", err)
	}
}

// TestStatsVerifyAndRebuildDetectsCounterDrift verifies repair behavior.
func TestStatsVerifyAndRebuildDetectsCounterDrift(t *testing.T) {
	db := newDB(t)
	repo := NewTicketRepository(orm.NewStore(db))
	ticket := validTicket()
	ticket.MessageCount = 99
	if _, err := repo.Create(context.Background(), ticket, validMessage(ticket.ID, "opener"), nil); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	report, err := repo.VerifyStats(context.Background())
	if err != nil {
		t.Fatalf("VerifyStats() error = %v", err)
	}
	if len(report.Mismatches) == 0 {
		t.Fatalf("Mismatches = 0, want drift")
	}
	report, err = repo.RebuildStats(context.Background())
	if err != nil {
		t.Fatalf("RebuildStats() error = %v", err)
	}
	if !report.Repaired {
		t.Fatalf("Repaired = false, want true")
	}
	after, err := repo.VerifyStats(context.Background())
	if err != nil {
		t.Fatalf("VerifyStats() after error = %v", err)
	}
	if len(after.Mismatches) != 0 {
		t.Fatalf("Mismatches after rebuild = %+v, want none", after.Mismatches)
	}
}

// TestTicketRepositoryUpdateFilterAndOperationalQueries verifies queue behavior.
func TestTicketRepositoryUpdateFilterAndOperationalQueries(t *testing.T) {
	db := newDB(t)
	repo := NewTicketRepository(orm.NewStore(db))
	ticket := validTicket()
	teamID := uuid.New()
	assigneeID := uuid.New()
	due := time.Now().UTC().Add(-time.Hour)
	ticket.CurrentTeamGroupID = &teamID
	ticket.AssigneeUserID = &assigneeID
	ticket.SLAResolutionDueAt = &due
	created, err := repo.Create(context.Background(), ticket, validMessage(ticket.ID, "opener"), nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	found, err := repo.FindByIdempotencyKey(context.Background(), ticket.Key)
	if err != nil {
		t.Fatalf("FindByIdempotencyKey() error = %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("found ID = %s, want %s", found.ID, created.ID)
	}
	created.Status = domain.StatusPendingStaff
	updated, err := repo.Update(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if _, err := repo.Update(context.Background(), updated, created.Version); err != port.ErrPreconditionFailed {
		t.Fatalf("stale Update() error = %v, want precondition", err)
	}
	result, err := repo.List(context.Background(), port.TicketFilter{CurrentTeamGroupID: teamID}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("filtered tickets = %d, want 1", len(result.Items))
	}
	breaches, err := repo.DetectSLABreaches(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("DetectSLABreaches() error = %v", err)
	}
	if len(breaches) != 1 {
		t.Fatalf("breaches = %d, want 1", len(breaches))
	}
	updated.Status = domain.StatusPendingSubmitter
	updated.UpdatedAt = time.Now().UTC().Add(-20 * 24 * time.Hour)
	model := ticketModel(updated)
	if err := db.Model(&TicketModel{}).Where("id = ?", updated.ID).Updates(map[string]any{
		"status":     model.Status,
		"updated_at": model.UpdatedAt,
		"version":    updated.Version,
	}).Error; err != nil {
		t.Fatalf("force stale update error = %v", err)
	}
	closed, err := repo.CloseStale(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("CloseStale() error = %v", err)
	}
	if closed != 1 {
		t.Fatalf("closed = %d, want 1", closed)
	}
}

// TestTicketRepositoryListPagesAreBounded verifies ticket queues and child timelines page safely.
func TestTicketRepositoryListPagesAreBounded(t *testing.T) {
	db := newDB(t)
	repo := NewTicketRepository(orm.NewStore(db))
	for index := 0; index < 2; index++ {
		ticket := validTicket()
		if _, err := repo.Create(context.Background(), ticket, validMessage(ticket.ID, "opener"), nil); err != nil {
			t.Fatalf("Create(%d) error = %v", index, err)
		}
	}
	tickets, err := repo.List(context.Background(), port.TicketFilter{}, pagination.Page{Limit: 1})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tickets.Items) != 1 || tickets.NextCursor == "" {
		t.Fatalf("tickets page = %+v, want one item and next cursor", tickets)
	}

	ticket := validTicket()
	if _, err := repo.Create(context.Background(), ticket, validMessage(ticket.ID, "opener"), nil); err != nil {
		t.Fatalf("Create message ticket error = %v", err)
	}
	if _, err := repo.AddMessage(context.Background(), validMessage(ticket.ID, "reply")); err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}
	messages, err := repo.ListMessages(context.Background(), ticket.ID, false, pagination.Page{Limit: 1})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages.Items) != 1 || messages.NextCursor == "" {
		t.Fatalf("messages page = %+v, want one item and next cursor", messages)
	}

	for index := 0; index < 2; index++ {
		action := testAction(ticket.ID)
		action.IdempotencyKey = uuid.NewString()
		if _, err := repo.AddAction(context.Background(), action); err != nil {
			t.Fatalf("AddAction(%d) error = %v", index, err)
		}
	}
	actions, err := repo.ListActions(context.Background(), ticket.ID, pagination.Page{Limit: 1})
	if err != nil {
		t.Fatalf("ListActions() error = %v", err)
	}
	if len(actions.Items) != 1 || actions.NextCursor == "" {
		t.Fatalf("actions page = %+v, want one item and next cursor", actions)
	}
}

// newDB creates an in-memory database.
func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&DefinitionModel{},
		&TicketModel{},
		&MessageModel{},
		&EvidenceModel{},
		&ActionModel{},
	); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

// validDefinition returns a persisted-ready definition.
func validDefinition() domain.Definition {
	return domain.Definition{
		ID:      uuid.New(),
		Key:     "support",
		Name:    "Support",
		Kind:    domain.KindSupport,
		Status:  domain.DefinitionActive,
		Version: 1,
	}
}

// validTicket returns a persisted-ready ticket.
func validTicket() domain.Ticket {
	now := time.Now().UTC()
	return domain.Ticket{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		Key:          uuid.NewString(),
		Title:        "Need help",
		Kind:         domain.KindSupport,
		Status:       domain.StatusOpen,
		Priority:     domain.PriorityNormal,
		OpenedAt:     now,
		Version:      1,
	}
}

// validMessage returns a persisted-ready message.
func validMessage(ticketID uuid.UUID, text string) domain.Message {
	return domain.Message{
		ID:                  uuid.New(),
		TicketID:            ticketID,
		AuthorRole:          domain.RoleSubmitter,
		Visibility:          domain.VisibilityParticipants,
		ContentFormat:       "prosemirror_json",
		ContentDocumentJSON: []byte(`{"type":"doc"}`),
		ContentText:         text,
		Version:             1,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
}

// testAction returns a persisted-ready action.
func testAction(ticketID uuid.UUID) domain.Action {
	return domain.Action{
		ID:          uuid.New(),
		TicketID:    ticketID,
		Type:        domain.ActionEscalate,
		Status:      domain.ActionCompleted,
		PayloadJSON: []byte(`{}`),
		CreatedAt:   time.Now().UTC(),
	}
}
