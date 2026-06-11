package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	"github.com/niflaot/gamehub-go/module/punishments/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// GetPunishment returns one punishment.
func (service Service) GetPunishment(ctx context.Context, id uuid.UUID) (domain.Punishment, error) {
	return service.cases.FindByID(ctx, id)
}

// ListPunishments returns matching punishments.
func (service Service) ListPunishments(ctx context.Context, filter port.PunishmentFilter, page pagination.Page) (pagination.Result[domain.Punishment], error) {
	return service.cases.List(ctx, filter, page)
}

// CheckRestriction denies when an active restriction matches.
func (service Service) CheckRestriction(ctx context.Context, command port.CheckCommand) (domain.CheckResult, error) {
	if command.UserID == uuid.Nil {
		return domain.CheckResult{Allowed: true}, nil
	}
	if service.cache != nil {
		if result, ok, err := service.cache.Get(ctx, command.UserID, command.ActionKey); err == nil && ok {
			return result, nil
		}
	}
	restriction, summary, err := service.cases.ActiveRestriction(
		ctx,
		command.UserID,
		command.ActionKey,
		time.Now().UTC(),
	)
	if err != nil {
		if err == port.ErrNotFound {
			result := domain.CheckResult{Allowed: true}
			service.cacheResult(ctx, command, result)
			return result, nil
		}
		return domain.CheckResult{}, err
	}
	result := domain.CheckResult{
		Allowed:     false,
		Punishment:  summary,
		Restriction: &restriction,
	}
	service.cacheResult(ctx, command, result)
	return result, nil
}

// Restricted reports whether userID is denied actionKey.
func (service Service) Restricted(ctx context.Context, userID uuid.UUID, actionKey string) (bool, error) {
	result, err := service.CheckRestriction(ctx, port.CheckCommand{UserID: userID, ActionKey: actionKey})
	if err != nil {
		return false, err
	}
	return !result.Allowed, nil
}

// ListActiveRestrictions returns active restrictions for a user.
func (service Service) ListActiveRestrictions(ctx context.Context, userID uuid.UUID) ([]domain.ActiveRestriction, error) {
	return service.cases.ListActiveRestrictions(ctx, userID, time.Now().UTC())
}

// ExpirePunishments expires due punishments.
func (service Service) ExpirePunishments(ctx context.Context) (int64, error) {
	count, err := service.cases.ExpireDue(ctx, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	if count > 0 {
		_ = service.ClearRestrictionCache(ctx)
		_ = service.publishOperationsEvent(ctx, "punishments.punishments.expired", count)
	}
	return count, nil
}

// VerifyRestrictions reports projection drift.
func (service Service) VerifyRestrictions(ctx context.Context) (domain.DriftReport, error) {
	return service.cases.VerifyRestrictions(ctx, time.Now().UTC())
}

// RebuildRestrictions repairs projection drift.
func (service Service) RebuildRestrictions(ctx context.Context) (domain.DriftReport, error) {
	report, err := service.cases.RebuildRestrictions(ctx, time.Now().UTC())
	if err != nil {
		return domain.DriftReport{}, err
	}
	report.Repaired = true
	_ = service.ClearRestrictionCache(ctx)
	_ = service.publishOperationsEvent(ctx, "punishments.restrictions.rebuilt", int64(len(report.Mismatches)))
	return report, nil
}

// ClearRestrictionCache clears cached restriction checks.
func (service Service) ClearRestrictionCache(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearAll(ctx)
}

func (service Service) prepareIssue(
	command port.IssueCommand,
	definition domain.Definition,
) (domain.Punishment, []domain.ActiveRestriction, error) {
	startsAt := command.StartsAt
	if startsAt.IsZero() {
		startsAt = time.Now().UTC()
	}
	if err := domain.ValidateIssueDuration(definition, startsAt, command.ExpiresAt); err != nil {
		return domain.Punishment{}, nil, err
	}
	if definition.RequiresReason && command.Reason == "" {
		return domain.Punishment{}, nil, domain.NewValidationError([]domain.Violation{
			{Field: "reason", Message: "is required"},
		})
	}
	if definition.RequiresTargetIP && command.TargetIPHash == "" {
		return domain.Punishment{}, nil, domain.NewValidationError([]domain.Violation{
			{Field: "target_ip_hash", Message: "is required"},
		})
	}
	punishment := domain.Punishment{
		ID:             uuid.New(),
		DefinitionID:   definition.ID,
		TargetUserID:   command.TargetUserID,
		TargetIPHash:   command.TargetIPHash,
		IssuerType:     command.IssuerType,
		IssuerUserID:   command.IssuerUserID,
		IssuerKey:      command.IssuerKey,
		Reason:         command.Reason,
		PrivateReason:  command.PrivateReason,
		Status:         domain.PunishmentActive,
		StartsAt:       startsAt,
		ExpiresAt:      command.ExpiresAt,
		Source:         command.Source,
		IdempotencyKey: command.IdempotencyKey,
		Version:        1,
	}.Normalize()
	if err := punishment.Validate(); err != nil {
		return domain.Punishment{}, nil, err
	}
	restrictions := []domain.ActiveRestriction{}
	for _, action := range definition.Actions {
		snapshot := domain.SnapshotFromTemplate(punishment.ID, action.Normalize())
		punishment.Snapshots = append(punishment.Snapshots, snapshot)
		if restriction, ok := domain.RestrictionFromSnapshot(punishment, snapshot); ok {
			restrictions = append(restrictions, restriction)
		}
	}
	return punishment, restrictions, nil
}

func (service Service) withinTx(ctx context.Context, fn func(context.Context) error) error {
	if service.transactions == nil {
		return fn(ctx)
	}
	return service.transactions.WithinTx(ctx, fn)
}

func (service Service) clearUser(ctx context.Context, userID uuid.UUID) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearUser(ctx, userID)
}

func (service Service) cacheResult(ctx context.Context, command port.CheckCommand, result domain.CheckResult) {
	if service.cache != nil {
		_ = service.cache.Set(ctx, command.UserID, command.ActionKey, result, restrictionCacheTTL)
	}
}

// Ensure Service implements port.Service.
var _ port.Service = Service{}
