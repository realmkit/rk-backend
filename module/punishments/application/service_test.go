package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestIssuePunishmentCreatesRestrictionAndEvent verifies issue orchestration.
func TestIssuePunishmentCreatesRestrictionAndEvent(t *testing.T) {
	definitions := newDefinitionFake(testDefinition())
	cases := newCaseFake()
	cache := &cacheFake{}
	events := &eventFake{}
	service := NewService(Dependencies{
		Definitions: definitions,
		Cases:       cases,
		Cache:       cache,
		Events:      events,
	})

	issued, err := service.IssuePunishment(context.Background(), port.IssueCommand{
		DefinitionID:   definitions.definition.ID,
		TargetUserID:   uuid.New(),
		IssuerType:     domain.IssuerSystem,
		IssuerKey:      "anticheat",
		Reason:         "spam",
		IdempotencyKey: "issue-1",
	})
	if err != nil {
		t.Fatalf("IssuePunishment() error = %v", err)
	}
	if len(issued.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want 1", len(issued.Snapshots))
	}
	if got := len(cases.restrictions); got != 1 {
		t.Fatalf("restrictions = %d, want 1", got)
	}
	if got := events.types[0]; got != "punishments.punishment.issued" {
		t.Fatalf("event type = %q, want issued", got)
	}
}

// TestCheckRestrictionUsesCache verifies cached restriction checks short-circuit storage.
func TestCheckRestrictionUsesCache(t *testing.T) {
	cache := &cacheFake{
		values: map[string]domain.CheckResult{
			domain.ActionForumsReply: {Allowed: false},
		},
	}
	service := NewService(Dependencies{
		Definitions: newDefinitionFake(testDefinition()),
		Cases:       newCaseFake(),
		Cache:       cache,
	})

	result, err := service.CheckRestriction(context.Background(), port.CheckCommand{
		UserID:    uuid.New(),
		ActionKey: domain.ActionForumsReply,
	})
	if err != nil {
		t.Fatalf("CheckRestriction() error = %v", err)
	}
	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}
}

// TestRevokePunishmentClearsRestriction verifies revoke removes active projections.
func TestRevokePunishmentClearsRestriction(t *testing.T) {
	cases := newCaseFake()
	service := NewService(Dependencies{
		Definitions: newDefinitionFake(testDefinition()),
		Cases:       cases,
		Cache:       &cacheFake{},
	})
	issued, err := service.IssuePunishment(context.Background(), port.IssueCommand{
		DefinitionID: cases.definitionID,
		TargetUserID: uuid.New(),
		IssuerType:   domain.IssuerSystem,
		IssuerKey:    "system",
		Reason:       "spam",
	})
	if err != nil {
		t.Fatalf("IssuePunishment() error = %v", err)
	}

	err = service.RevokePunishment(context.Background(), port.RevokeCommand{
		ActorUserID:     uuid.New(),
		PunishmentID:    issued.ID,
		Reason:          "appeal",
		ExpectedVersion: issued.Version,
	})
	if err != nil {
		t.Fatalf("RevokePunishment() error = %v", err)
	}
	if len(cases.restrictions) != 0 {
		t.Fatalf("restrictions = %d, want 0", len(cases.restrictions))
	}
}

// TestIssuePunishmentRequiresDefinitionTargetIP verifies issue-time definition policy.
func TestIssuePunishmentRequiresDefinitionTargetIP(t *testing.T) {
	definition := testDefinition()
	definition.RequiresTargetIP = true
	service := NewService(Dependencies{
		Definitions: newDefinitionFake(definition),
		Cases:       newCaseFake(),
	})

	_, err := service.IssuePunishment(context.Background(), port.IssueCommand{
		DefinitionID: definition.ID,
		TargetUserID: uuid.New(),
		IssuerType:   domain.IssuerSystem,
		IssuerKey:    "system",
		Reason:       "spam",
	})
	if err == nil {
		t.Fatalf("IssuePunishment() error = nil, want validation error")
	}
}

type definitionFake struct {
	definition domain.Definition
	deletedID  uuid.UUID
	reordered  []uuid.UUID
}

func newDefinitionFake(definition domain.Definition) *definitionFake {
	return &definitionFake{definition: definition}
}

func (fake *definitionFake) Create(_ context.Context, definition domain.Definition) (domain.Definition, error) {
	fake.definition = definition
	return fake.definition, nil
}

func (fake *definitionFake) Update(_ context.Context, definition domain.Definition, _ uint64) (domain.Definition, error) {
	fake.definition = definition
	return fake.definition, nil
}

func (fake *definitionFake) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	fake.deletedID = id
	return nil
}

func (fake *definitionFake) FindByID(context.Context, uuid.UUID) (domain.Definition, error) {
	return fake.definition, nil
}

func (fake *definitionFake) List(context.Context, port.DefinitionFilter, pagination.Page) (pagination.Result[domain.Definition], error) {
	return pagination.Result[domain.Definition]{Items: []domain.Definition{fake.definition}}, nil
}

func (fake *definitionFake) ReorderActions(_ context.Context, _ uuid.UUID, actionIDs []uuid.UUID) error {
	fake.reordered = actionIDs
	return nil
}

type caseFake struct {
	definitionID  uuid.UUID
	punishments   map[uuid.UUID]domain.Punishment
	idempotency   map[string]uuid.UUID
	restrictions  []domain.ActiveRestriction
	expireCount   int64
	verifyReport  domain.DriftReport
	rebuildReport domain.DriftReport
}

func newCaseFake() *caseFake {
	return &caseFake{
		definitionID: uuid.New(),
		punishments:  map[uuid.UUID]domain.Punishment{},
		idempotency:  map[string]uuid.UUID{},
	}
}

func (fake *caseFake) Issue(
	_ context.Context,
	punishment domain.Punishment,
	restrictions []domain.ActiveRestriction,
) (domain.Punishment, error) {
	fake.punishments[punishment.ID] = punishment
	if punishment.IdempotencyKey != "" {
		fake.idempotency[punishment.IdempotencyKey] = punishment.ID
	}
	fake.restrictions = append(fake.restrictions, restrictions...)
	return punishment, nil
}

func (fake *caseFake) Update(_ context.Context, punishment domain.Punishment, _ uint64) (domain.Punishment, error) {
	fake.punishments[punishment.ID] = punishment
	return punishment, nil
}

func (fake *caseFake) Revoke(_ context.Context, punishment domain.Punishment, _ uint64) error {
	fake.punishments[punishment.ID] = punishment
	fake.restrictions = nil
	return nil
}

func (fake *caseFake) ExpireDue(context.Context, time.Time) (int64, error) {
	return fake.expireCount, nil
}

func (fake *caseFake) FindByID(_ context.Context, id uuid.UUID) (domain.Punishment, error) {
	punishment, ok := fake.punishments[id]
	if !ok {
		return domain.Punishment{}, port.ErrNotFound
	}
	return punishment, nil
}

func (fake *caseFake) FindByIdempotencyKey(_ context.Context, key string) (domain.Punishment, error) {
	id, ok := fake.idempotency[key]
	if !ok {
		return domain.Punishment{}, port.ErrNotFound
	}
	return fake.punishments[id], nil
}

func (fake *caseFake) List(
	context.Context,
	port.PunishmentFilter,
	pagination.Page,
) (pagination.Result[domain.Punishment], error) {
	items := make([]domain.Punishment, 0, len(fake.punishments))
	for _, punishment := range fake.punishments {
		items = append(items, punishment)
	}
	return pagination.Result[domain.Punishment]{Items: items}, nil
}

func (fake *caseFake) ActiveRestriction(
	_ context.Context,
	_ uuid.UUID,
	actionKey string,
	_ time.Time,
) (domain.ActiveRestriction, *domain.PunishmentSummary, error) {
	for _, restriction := range fake.restrictions {
		if restriction.ActionKey == actionKey {
			return restriction, &domain.PunishmentSummary{ID: restriction.PunishmentID}, nil
		}
	}
	return domain.ActiveRestriction{}, nil, port.ErrNotFound
}

func (fake *caseFake) ListActiveRestrictions(context.Context, uuid.UUID, time.Time) ([]domain.ActiveRestriction, error) {
	return fake.restrictions, nil
}

func (fake *caseFake) VerifyRestrictions(context.Context, time.Time) (domain.DriftReport, error) {
	return fake.verifyReport, nil
}

func (fake *caseFake) RebuildRestrictions(context.Context, time.Time) (domain.DriftReport, error) {
	if len(fake.rebuildReport.Mismatches) > 0 {
		return fake.rebuildReport, nil
	}
	return domain.DriftReport{Repaired: true}, nil
}

type cacheFake struct {
	values       map[string]domain.CheckResult
	clearedUsers []uuid.UUID
	clearAllHits int
}

func (fake *cacheFake) Get(_ context.Context, _ uuid.UUID, actionKey string) (domain.CheckResult, bool, error) {
	result, ok := fake.values[actionKey]
	return result, ok, nil
}

func (fake *cacheFake) Set(_ context.Context, _ uuid.UUID, actionKey string, result domain.CheckResult, _ time.Duration) error {
	if fake.values == nil {
		fake.values = map[string]domain.CheckResult{}
	}
	fake.values[actionKey] = result
	return nil
}

func (fake *cacheFake) ClearUser(_ context.Context, userID uuid.UUID) error {
	fake.clearedUsers = append(fake.clearedUsers, userID)
	return nil
}

func (fake *cacheFake) ClearAll(context.Context) error {
	fake.clearAllHits++
	return nil
}

type eventFake struct {
	types  []string
	drafts []eventdomain.Draft
}

func (fake *eventFake) Publish(_ context.Context, draft eventdomain.Draft) (eventdomain.Event, error) {
	fake.types = append(fake.types, string(draft.Key))
	fake.drafts = append(fake.drafts, draft)
	return eventdomain.Event{ID: uuid.New(), Key: draft.Key}, nil
}

func testDefinition() domain.Definition {
	action := domain.ActionTemplate{
		ID:                uuid.New(),
		TargetSystem:      domain.TargetRealmKit,
		ActionType:        domain.ActionForumsReply,
		ConfigurationJSON: []byte(`{}`),
		Status:            domain.DefinitionActive,
	}.Normalize()
	return domain.Definition{
		ID:             uuid.New(),
		Key:            "chat_ban",
		Name:           "Chat Ban",
		Color:          "#ff5555",
		Status:         domain.DefinitionActive,
		AllowPermanent: true,
		RequiresReason: true,
		Actions:        []domain.ActionTemplate{action},
		Version:        1,
	}.Normalize()
}
