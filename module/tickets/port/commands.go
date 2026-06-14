package port

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/pkg/search"
)

// DefinitionFilter filters ticket definition lists.
type DefinitionFilter struct {
	// Kind filters by ticket nature.
	Kind domain.Kind

	// Status filters by definition lifecycle state.
	Status domain.DefinitionStatus

	// Query filters by key, name, or description.
	Query search.TextQuery

	// Sort controls deterministic result ordering.
	Sort search.Sort
}

// TicketFilter filters ticket lists.
type TicketFilter struct {
	// SubmitterUserID filters by submitter.
	SubmitterUserID uuid.UUID
	// TargetUserID filters by target user.
	TargetUserID uuid.UUID
	// PunishmentID filters by linked punishment.
	PunishmentID uuid.UUID
	// CurrentTeamGroupID filters by current team queue.
	CurrentTeamGroupID uuid.UUID
	// AssigneeUserID filters by assignee.
	AssigneeUserID uuid.UUID
	// Status filters by ticket lifecycle status.
	Status domain.TicketStatus
	// Kind filters by ticket nature.
	Kind domain.Kind
	// Query filters by title and stable identifiers.
	Query search.TextQuery
	// Sort controls deterministic result ordering.
	Sort search.Sort
}

// DefaultDefinitionSort returns the default ticket definition sort.
func DefaultDefinitionSort() search.SortOption {
	return search.SortOption{Key: "display_order", DefaultDirection: search.DirectionAsc}
}

// AllowedDefinitionSorts returns public ticket definition sort keys.
func AllowedDefinitionSorts() []search.SortOption {
	return []search.SortOption{
		DefaultDefinitionSort(),
		{Key: "name", DefaultDirection: search.DirectionAsc},
		{Key: "created_at", DefaultDirection: search.DirectionDesc},
	}
}

// DefaultTicketSort returns the default ticket queue sort.
func DefaultTicketSort() search.SortOption {
	return search.SortOption{Key: "updated_at", DefaultDirection: search.DirectionDesc}
}

// AllowedTicketSorts returns public ticket queue sort keys.
func AllowedTicketSorts() []search.SortOption {
	return []search.SortOption{
		DefaultTicketSort(),
		{Key: "created_at", DefaultDirection: search.DirectionDesc},
		{Key: "priority", DefaultDirection: search.DirectionDesc},
		{Key: "title", DefaultDirection: search.DirectionAsc},
	}
}

// CreateTicketCommand opens one ticket.
type CreateTicketCommand struct {
	ActorUserID         uuid.UUID
	DefinitionID        uuid.UUID
	Title               string
	SubmitterUserID     *uuid.UUID
	TargetUserID        *uuid.UUID
	PunishmentID        *uuid.UUID
	ContentDocumentJSON json.RawMessage
	ContentText         string
	EvidenceAssetIDs    []uuid.UUID
	IdempotencyKey      string
}

// MessageCommand creates a ticket message.
type MessageCommand struct {
	ActorUserID         uuid.UUID
	TicketID            uuid.UUID
	Visibility          domain.MessageVisibility
	ContentDocumentJSON json.RawMessage
	ContentText         string
	IdempotencyKey      string
}

// EvidenceCommand adds ticket evidence.
type EvidenceCommand struct {
	ActorUserID    uuid.UUID
	TicketID       uuid.UUID
	MessageID      *uuid.UUID
	AssetID        *uuid.UUID
	ExternalURL    string
	Label          string
	Description    string
	Visibility     domain.MessageVisibility
	IdempotencyKey string
}

// StaffCommand changes ticket workflow state.
type StaffCommand struct {
	ActorUserID     uuid.UUID
	TicketID        uuid.UUID
	AssigneeUserID  *uuid.UUID
	TeamGroupID     *uuid.UUID
	Reason          string
	ExpectedVersion uint64
	IdempotencyKey  string
}

// AppealDecisionCommand accepts or rejects an appeal.
type AppealDecisionCommand struct {
	ActorUserID      uuid.UUID
	TicketID         uuid.UUID
	Reason           string
	RevokePunishment bool
	ExpectedVersion  uint64
	IdempotencyKey   string
}
