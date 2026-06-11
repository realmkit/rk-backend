// Package application coordinates punishment use cases.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

const restrictionCacheTTL = 30 * time.Second

// Dependencies contains punishment service dependencies.
type Dependencies struct {
	Definitions  port.DefinitionRepository
	Cases        port.CaseRepository
	Cache        port.RestrictionCache
	Transactions port.TransactionRunner
	Events       emitter.Publisher
}

// Service implements punishment use cases.
type Service struct {
	definitions  port.DefinitionRepository
	cases        port.CaseRepository
	cache        port.RestrictionCache
	transactions port.TransactionRunner
	events       emitter.Publisher
}

// NewService creates a punishment service.
func NewService(deps Dependencies) Service {
	return Service{
		definitions:  deps.Definitions,
		cases:        deps.Cases,
		cache:        deps.Cache,
		transactions: deps.Transactions,
		events:       deps.Events,
	}
}

// CreateDefinition stores a punishment definition.
func (service Service) CreateDefinition(ctx context.Context, definition domain.Definition) (domain.Definition, error) {
	definition = definition.Normalize()
	for index := range definition.Actions {
		definition.Actions[index] = definition.Actions[index].Normalize()
	}
	if err := definition.Validate(); err != nil {
		return domain.Definition{}, err
	}
	created, err := service.definitions.Create(ctx, definition)
	if err != nil {
		return domain.Definition{}, err
	}
	return created, service.publishDefinitionEvent(ctx, "punishments.definition.created", created)
}

// UpdateDefinition updates a punishment definition.
func (service Service) UpdateDefinition(
	ctx context.Context,
	definition domain.Definition,
	expectedVersion uint64,
) (domain.Definition, error) {
	definition = definition.Normalize()
	if err := definition.Validate(); err != nil {
		return domain.Definition{}, err
	}
	updated, err := service.definitions.Update(ctx, definition, expectedVersion)
	if err != nil {
		return domain.Definition{}, err
	}
	return updated, service.publishDefinitionEvent(ctx, "punishments.definition.updated", updated)
}

// DeleteDefinition soft deletes one definition.
func (service Service) DeleteDefinition(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	current, err := service.definitions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := service.definitions.Delete(ctx, id, expectedVersion); err != nil {
		return err
	}
	return service.publishDefinitionEvent(ctx, "punishments.definition.deleted", current)
}

// GetDefinition returns one definition.
func (service Service) GetDefinition(ctx context.Context, id uuid.UUID) (domain.Definition, error) {
	return service.definitions.FindByID(ctx, id)
}

// ListDefinitions returns definitions.
func (service Service) ListDefinitions(
	ctx context.Context,
	filter port.DefinitionFilter,
	page pagination.Page,
) (pagination.Result[domain.Definition], error) {
	return service.definitions.List(ctx, filter, page)
}

// ReorderDefinitionActions reorders definition actions.
func (service Service) ReorderDefinitionActions(ctx context.Context, definitionID uuid.UUID, actionIDs []uuid.UUID) error {
	if err := service.definitions.ReorderActions(ctx, definitionID, actionIDs); err != nil {
		return err
	}
	definition, err := service.definitions.FindByID(ctx, definitionID)
	if err != nil {
		return err
	}
	return service.publishDefinitionEvent(ctx, "punishments.definition.updated", definition)
}

// IssuePunishment creates a punishment, snapshots, restrictions, and events.
func (service Service) IssuePunishment(ctx context.Context, command port.IssueCommand) (domain.Punishment, error) {
	if command.IdempotencyKey != "" {
		if existing, err := service.cases.FindByIdempotencyKey(ctx, command.IdempotencyKey); err == nil {
			return existing, nil
		}
	}
	definition, err := service.definitions.FindByID(ctx, command.DefinitionID)
	if err != nil {
		return domain.Punishment{}, err
	}
	if definition.Status != domain.DefinitionActive {
		return domain.Punishment{}, port.ErrConflict
	}
	punishment, restrictions, err := service.prepareIssue(command, definition)
	if err != nil {
		return domain.Punishment{}, err
	}
	var issued domain.Punishment
	err = service.withinTx(ctx, func(ctx context.Context) error {
		stored, err := service.cases.Issue(ctx, punishment, restrictions)
		if err != nil {
			return err
		}
		issued = stored
		if err := service.clearUser(ctx, stored.TargetUserID); err != nil {
			return err
		}
		return service.publishPunishmentEvent(ctx, "punishments.punishment.issued", stored)
	})
	return issued, err
}

// UpdatePunishment updates visible punishment notes.
func (service Service) UpdatePunishment(ctx context.Context, command port.UpdateCommand) (domain.Punishment, error) {
	current, err := service.cases.FindByID(ctx, command.PunishmentID)
	if err != nil {
		return domain.Punishment{}, err
	}
	current.Reason = command.Reason
	current.PrivateReason = command.PrivateReason
	current = current.Normalize()
	if err := current.Validate(); err != nil {
		return domain.Punishment{}, err
	}
	updated, err := service.cases.Update(ctx, current, command.ExpectedVersion)
	if err != nil {
		return domain.Punishment{}, err
	}
	return updated, service.publishPunishmentEvent(ctx, "punishments.punishment.updated", updated)
}

// RevokePunishment revokes an active punishment and clears restrictions.
func (service Service) RevokePunishment(ctx context.Context, command port.RevokeCommand) error {
	current, err := service.cases.FindByID(ctx, command.PunishmentID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	current.Status = domain.PunishmentRevoked
	current.RevokedAt = &now
	current.RevokedByUserID = &command.ActorUserID
	current.RevocationReason = command.Reason
	return service.withinTx(ctx, func(ctx context.Context) error {
		if err := service.cases.Revoke(ctx, current, command.ExpectedVersion); err != nil {
			return err
		}
		if err := service.clearUser(ctx, current.TargetUserID); err != nil {
			return err
		}
		return service.publishPunishmentEvent(ctx, "punishments.punishment.revoked", current)
	})
}
