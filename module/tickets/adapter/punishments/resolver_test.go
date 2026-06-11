package punishments

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	punishmentdomain "github.com/niflaot/gamehub-go/module/punishments/domain"
	punishmentport "github.com/niflaot/gamehub-go/module/punishments/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestResolverMapsAndRevokesPunishments verifies appeal integration.
func TestResolverMapsAndRevokesPunishments(t *testing.T) {
	service := &fakePunishmentService{punishment: punishmentdomain.Punishment{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		TargetUserID: uuid.New(),
		Status:       punishmentdomain.PunishmentActive,
		StartsAt:     time.Now().UTC(),
		Version:      5,
	}}
	resolver := NewResolver(service)
	summary, err := resolver.GetPunishment(context.Background(), service.punishment.ID)
	if err != nil {
		t.Fatalf("GetPunishment() error = %v", err)
	}
	if summary.TargetUserID != service.punishment.TargetUserID {
		t.Fatalf("TargetUserID = %s, want %s", summary.TargetUserID, service.punishment.TargetUserID)
	}
	actor := uuid.New()
	if err := resolver.RevokePunishment(context.Background(), service.punishment.ID, actor, "appeal accepted", 0); err != nil {
		t.Fatalf("RevokePunishment() error = %v", err)
	}
	if service.revoke.ExpectedVersion != 5 {
		t.Fatalf("ExpectedVersion = %d, want 5", service.revoke.ExpectedVersion)
	}
}

// fakePunishmentService implements punishment service for bridge tests.
type fakePunishmentService struct {
	punishment punishmentdomain.Punishment
	revoke     punishmentport.RevokeCommand
}

// CreateDefinition is unused.
func (fake *fakePunishmentService) CreateDefinition(context.Context, punishmentdomain.Definition) (punishmentdomain.Definition, error) {
	return punishmentdomain.Definition{}, nil
}

// UpdateDefinition is unused.
func (fake *fakePunishmentService) UpdateDefinition(
	context.Context,
	punishmentdomain.Definition,
	uint64,
) (punishmentdomain.Definition, error) {
	return punishmentdomain.Definition{}, nil
}

// DeleteDefinition is unused.
func (fake *fakePunishmentService) DeleteDefinition(context.Context, uuid.UUID, uint64) error {
	return nil
}

// GetDefinition is unused.
func (fake *fakePunishmentService) GetDefinition(context.Context, uuid.UUID) (punishmentdomain.Definition, error) {
	return punishmentdomain.Definition{}, nil
}

// ListDefinitions is unused.
func (fake *fakePunishmentService) ListDefinitions(
	context.Context,
	punishmentport.DefinitionFilter,
	pagination.Page,
) (pagination.Result[punishmentdomain.Definition], error) {
	return pagination.Result[punishmentdomain.Definition]{}, nil
}

// ReorderDefinitionActions is unused.
func (fake *fakePunishmentService) ReorderDefinitionActions(context.Context, uuid.UUID, []uuid.UUID) error {
	return nil
}

// IssuePunishment is unused.
func (fake *fakePunishmentService) IssuePunishment(context.Context, punishmentport.IssueCommand) (punishmentdomain.Punishment, error) {
	return punishmentdomain.Punishment{}, nil
}

// UpdatePunishment is unused.
func (fake *fakePunishmentService) UpdatePunishment(context.Context, punishmentport.UpdateCommand) (punishmentdomain.Punishment, error) {
	return punishmentdomain.Punishment{}, nil
}

// RevokePunishment records revoke command.
func (fake *fakePunishmentService) RevokePunishment(_ context.Context, command punishmentport.RevokeCommand) error {
	fake.revoke = command
	return nil
}

// GetPunishment returns fake punishment.
func (fake *fakePunishmentService) GetPunishment(context.Context, uuid.UUID) (punishmentdomain.Punishment, error) {
	return fake.punishment, nil
}

// ListPunishments is unused.
func (fake *fakePunishmentService) ListPunishments(
	context.Context,
	punishmentport.PunishmentFilter,
	pagination.Page,
) (pagination.Result[punishmentdomain.Punishment], error) {
	return pagination.Result[punishmentdomain.Punishment]{}, nil
}

// CheckRestriction is unused.
func (fake *fakePunishmentService) CheckRestriction(context.Context, punishmentport.CheckCommand) (punishmentdomain.CheckResult, error) {
	return punishmentdomain.CheckResult{}, nil
}

// ListActiveRestrictions is unused.
func (fake *fakePunishmentService) ListActiveRestrictions(context.Context, uuid.UUID) ([]punishmentdomain.ActiveRestriction, error) {
	return nil, nil
}

// ExpirePunishments is unused.
func (fake *fakePunishmentService) ExpirePunishments(context.Context) (int64, error) { return 0, nil }

// VerifyRestrictions is unused.
func (fake *fakePunishmentService) VerifyRestrictions(context.Context) (punishmentdomain.DriftReport, error) {
	return punishmentdomain.DriftReport{}, nil
}

// RebuildRestrictions is unused.
func (fake *fakePunishmentService) RebuildRestrictions(context.Context) (punishmentdomain.DriftReport, error) {
	return punishmentdomain.DriftReport{}, nil
}

// ClearRestrictionCache is unused.
func (fake *fakePunishmentService) ClearRestrictionCache(context.Context) error { return nil }
