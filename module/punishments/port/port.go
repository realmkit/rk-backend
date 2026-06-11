// Package port defines punishment application contracts.
package port

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

var (
	// ErrNotFound reports a missing punishment resource.
	ErrNotFound = errors.New("punishment not found")
	// ErrConflict reports conflicting punishment state.
	ErrConflict = errors.New("punishment conflict")
	// ErrForbidden reports denied access or restricted action.
	ErrForbidden = errors.New("punishment forbidden")
	// ErrPreconditionFailed reports stale optimistic version.
	ErrPreconditionFailed = errors.New("punishment precondition failed")
)

// DefinitionFilter filters definition lists.
type DefinitionFilter struct {
	Status domain.DefinitionStatus
}

// PunishmentFilter filters punishment lists.
type PunishmentFilter struct {
	TargetUserID uuid.UUID
	Status       domain.PunishmentStatus
}

// IssueCommand issues a punishment.
type IssueCommand struct {
	ActorUserID    uuid.UUID
	DefinitionID   uuid.UUID
	TargetUserID   uuid.UUID
	TargetIPHash   string
	IssuerType     domain.IssuerType
	IssuerUserID   *uuid.UUID
	IssuerKey      string
	Reason         string
	PrivateReason  string
	StartsAt       time.Time
	ExpiresAt      *time.Time
	Source         string
	IdempotencyKey string
}

// RevokeCommand revokes a punishment.
type RevokeCommand struct {
	ActorUserID     uuid.UUID
	PunishmentID    uuid.UUID
	Reason          string
	ExpectedVersion uint64
	IdempotencyKey  string
}

// UpdateCommand updates non-state punishment fields.
type UpdateCommand struct {
	ActorUserID     uuid.UUID
	PunishmentID    uuid.UUID
	Reason          string
	PrivateReason   string
	ExpectedVersion uint64
}

// CheckCommand checks whether an action is restricted.
type CheckCommand struct {
	UserID    uuid.UUID
	ActionKey string
}

// DefinitionRepository stores definitions and action templates.
type DefinitionRepository interface {
	Create(ctx context.Context, definition domain.Definition) (domain.Definition, error)
	Update(ctx context.Context, definition domain.Definition, expectedVersion uint64) (domain.Definition, error)
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error
	FindByID(ctx context.Context, id uuid.UUID) (domain.Definition, error)
	List(ctx context.Context, filter DefinitionFilter, page pagination.Page) (pagination.Result[domain.Definition], error)
	ReorderActions(ctx context.Context, definitionID uuid.UUID, actionIDs []uuid.UUID) error
}

// CaseRepository stores issued punishments and projections.
type CaseRepository interface {
	Issue(ctx context.Context, punishment domain.Punishment, restrictions []domain.ActiveRestriction) (domain.Punishment, error)
	Update(ctx context.Context, punishment domain.Punishment, expectedVersion uint64) (domain.Punishment, error)
	Revoke(ctx context.Context, punishment domain.Punishment, expectedVersion uint64) error
	ExpireDue(ctx context.Context, now time.Time) (int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.Punishment, error)
	FindByIdempotencyKey(ctx context.Context, key string) (domain.Punishment, error)
	List(ctx context.Context, filter PunishmentFilter, page pagination.Page) (pagination.Result[domain.Punishment], error)
	ActiveRestriction(ctx context.Context, userID uuid.UUID, actionKey string, now time.Time) (domain.ActiveRestriction, *domain.PunishmentSummary, error)
	ListActiveRestrictions(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.ActiveRestriction, error)
	VerifyRestrictions(ctx context.Context, now time.Time) (domain.DriftReport, error)
	RebuildRestrictions(ctx context.Context, now time.Time) (domain.DriftReport, error)
}

// RestrictionCache caches active restrictions.
type RestrictionCache interface {
	Get(ctx context.Context, userID uuid.UUID, actionKey string) (domain.CheckResult, bool, error)
	Set(ctx context.Context, userID uuid.UUID, actionKey string, result domain.CheckResult, ttl time.Duration) error
	ClearUser(ctx context.Context, userID uuid.UUID) error
	ClearAll(ctx context.Context) error
}

// TransactionRunner runs work in a transaction.
type TransactionRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Service exposes punishment use cases.
type Service interface {
	CreateDefinition(context.Context, domain.Definition) (domain.Definition, error)
	UpdateDefinition(context.Context, domain.Definition, uint64) (domain.Definition, error)
	DeleteDefinition(context.Context, uuid.UUID, uint64) error
	GetDefinition(context.Context, uuid.UUID) (domain.Definition, error)
	ListDefinitions(context.Context, DefinitionFilter, pagination.Page) (pagination.Result[domain.Definition], error)
	ReorderDefinitionActions(context.Context, uuid.UUID, []uuid.UUID) error
	IssuePunishment(context.Context, IssueCommand) (domain.Punishment, error)
	UpdatePunishment(context.Context, UpdateCommand) (domain.Punishment, error)
	RevokePunishment(context.Context, RevokeCommand) error
	GetPunishment(context.Context, uuid.UUID) (domain.Punishment, error)
	ListPunishments(context.Context, PunishmentFilter, pagination.Page) (pagination.Result[domain.Punishment], error)
	CheckRestriction(context.Context, CheckCommand) (domain.CheckResult, error)
	ListActiveRestrictions(context.Context, uuid.UUID) ([]domain.ActiveRestriction, error)
	ExpirePunishments(context.Context) (int64, error)
	VerifyRestrictions(context.Context) (domain.DriftReport, error)
	RebuildRestrictions(context.Context) (domain.DriftReport, error)
	ClearRestrictionCache(context.Context) error
}
