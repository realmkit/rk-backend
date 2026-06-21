// Package port defines punishment application contracts.
package port

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
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
	// Query filters by key, name, or description.
	Query search.TextQuery

	// Sort controls deterministic result ordering.
	Sort search.Sort

	// Status filters by lifecycle status.
	Status domain.DefinitionStatus
}

// PunishmentFilter filters punishment lists.
type PunishmentFilter struct {
	// TargetUserID filters by punished user.
	TargetUserID uuid.UUID

	// Status filters by punishment lifecycle status.
	Status domain.PunishmentStatus

	// Query filters by reason, source, issuer key, or target IP hash.
	Query search.TextQuery

	// Sort controls deterministic result ordering.
	Sort search.Sort
}

// DefaultDefinitionSort returns the default punishment definition sort.
func DefaultDefinitionSort() search.SortOption {
	return search.SortOption{Key: "display_order", DefaultDirection: search.DirectionAsc}
}

// AllowedDefinitionSorts returns public punishment definition list sort keys.
func AllowedDefinitionSorts() []search.SortOption {
	return []search.SortOption{
		DefaultDefinitionSort(),
		{Key: "name", DefaultDirection: search.DirectionAsc},
		{Key: "severity", DefaultDirection: search.DirectionDesc},
		{Key: "created_at", DefaultDirection: search.DirectionDesc},
	}
}

// DefaultPunishmentSort returns the default punishment case sort.
func DefaultPunishmentSort() search.SortOption {
	return search.SortOption{Key: "created_at", DefaultDirection: search.DirectionDesc}
}

// AllowedPunishmentSorts returns public punishment case list sort keys.
func AllowedPunishmentSorts() []search.SortOption {
	return []search.SortOption{
		DefaultPunishmentSort(),
		{Key: "expires_at", DefaultDirection: search.DirectionAsc},
		{Key: "status", DefaultDirection: search.DirectionAsc},
	}
}

// IssueCommand issues a punishment.
type IssueCommand struct {
	ActorUserID    uuid.UUID         // ActorUserID stores the actor user i d value.
	DefinitionID   uuid.UUID         // DefinitionID stores the definition i d value.
	TargetUserID   uuid.UUID         // TargetUserID stores the target user i d value.
	TargetIPHash   string            // TargetIPHash stores the target i p hash value.
	IssuerType     domain.IssuerType // IssuerType stores the issuer type value.
	IssuerUserID   *uuid.UUID        // IssuerUserID stores the issuer user i d value.
	IssuerKey      string            // IssuerKey stores the issuer key value.
	Reason         string            // Reason stores the reason value.
	PrivateReason  string            // PrivateReason stores the private reason value.
	StartsAt       time.Time         // StartsAt stores the starts at value.
	ExpiresAt      *time.Time        // ExpiresAt stores the expires at value.
	Source         string            // Source stores the source value.
	IdempotencyKey string            // IdempotencyKey stores the idempotency key value.
}

// RevokeCommand revokes a punishment.
type RevokeCommand struct {
	ActorUserID     uuid.UUID // ActorUserID stores the actor user i d value.
	PunishmentID    uuid.UUID // PunishmentID stores the punishment i d value.
	Reason          string    // Reason stores the reason value.
	ExpectedVersion uint64    // ExpectedVersion stores the expected version value.
	IdempotencyKey  string    // IdempotencyKey stores the idempotency key value.
}

// UpdateCommand updates non-state punishment fields.
type UpdateCommand struct {
	ActorUserID     uuid.UUID // ActorUserID stores the actor user i d value.
	PunishmentID    uuid.UUID // PunishmentID stores the punishment i d value.
	Reason          string    // Reason stores the reason value.
	PrivateReason   string    // PrivateReason stores the private reason value.
	ExpectedVersion uint64    // ExpectedVersion stores the expected version value.
}

// CheckCommand checks whether an action is restricted.
type CheckCommand struct {
	UserID    uuid.UUID // UserID stores the user i d value.
	ActionKey string    // ActionKey stores the action key value.
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
	ActiveRestriction(
		ctx context.Context,
		userID uuid.UUID,
		actionKey string,
		now time.Time,
	) (domain.ActiveRestriction, *domain.PunishmentSummary, error)
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
