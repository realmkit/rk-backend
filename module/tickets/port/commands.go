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
	ActorUserID         uuid.UUID       // ActorUserID stores the actor user i d value.
	DefinitionID        uuid.UUID       // DefinitionID stores the definition i d value.
	Title               string          // Title stores the title value.
	SubmitterUserID     *uuid.UUID      // SubmitterUserID stores the submitter user i d value.
	TargetUserID        *uuid.UUID      // TargetUserID stores the target user i d value.
	PunishmentID        *uuid.UUID      // PunishmentID stores the punishment i d value.
	ContentDocumentJSON json.RawMessage // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string          // ContentText stores the content text value.
	EvidenceAssetIDs    []uuid.UUID     // EvidenceAssetIDs stores the evidence asset i ds value.
	IdempotencyKey      string          // IdempotencyKey stores the idempotency key value.
}

// MessageCommand creates a ticket message.
type MessageCommand struct {
	ActorUserID         uuid.UUID                // ActorUserID stores the actor user i d value.
	TicketID            uuid.UUID                // TicketID stores the ticket i d value.
	Visibility          domain.MessageVisibility // Visibility stores the visibility value.
	ContentDocumentJSON json.RawMessage          // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string                   // ContentText stores the content text value.
	IdempotencyKey      string                   // IdempotencyKey stores the idempotency key value.
}

// EvidenceCommand adds ticket evidence.
type EvidenceCommand struct {
	ActorUserID    uuid.UUID                // ActorUserID stores the actor user i d value.
	TicketID       uuid.UUID                // TicketID stores the ticket i d value.
	MessageID      *uuid.UUID               // MessageID stores the message i d value.
	AssetID        *uuid.UUID               // AssetID stores the asset i d value.
	ExternalURL    string                   // ExternalURL stores the external u r l value.
	Label          string                   // Label stores the label value.
	Description    string                   // Description stores the description value.
	Visibility     domain.MessageVisibility // Visibility stores the visibility value.
	IdempotencyKey string                   // IdempotencyKey stores the idempotency key value.
}

// StaffCommand changes ticket workflow state.
type StaffCommand struct {
	ActorUserID     uuid.UUID  // ActorUserID stores the actor user i d value.
	TicketID        uuid.UUID  // TicketID stores the ticket i d value.
	AssigneeUserID  *uuid.UUID // AssigneeUserID stores the assignee user i d value.
	TeamGroupID     *uuid.UUID // TeamGroupID stores the team group i d value.
	Reason          string     // Reason stores the reason value.
	ExpectedVersion uint64     // ExpectedVersion stores the expected version value.
	IdempotencyKey  string     // IdempotencyKey stores the idempotency key value.
}

// AppealDecisionCommand accepts or rejects an appeal.
type AppealDecisionCommand struct {
	ActorUserID      uuid.UUID // ActorUserID stores the actor user i d value.
	TicketID         uuid.UUID // TicketID stores the ticket i d value.
	Reason           string    // Reason stores the reason value.
	RevokePunishment bool      // RevokePunishment stores the revoke punishment value.
	ExpectedVersion  uint64    // ExpectedVersion stores the expected version value.
	IdempotencyKey   string    // IdempotencyKey stores the idempotency key value.
}
