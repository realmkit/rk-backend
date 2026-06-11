package port

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
)

// DefinitionFilter filters ticket definition lists.
type DefinitionFilter struct {
	Kind   domain.Kind
	Status domain.DefinitionStatus
}

// TicketFilter filters ticket lists.
type TicketFilter struct {
	SubmitterUserID    uuid.UUID
	TargetUserID       uuid.UUID
	PunishmentID       uuid.UUID
	CurrentTeamGroupID uuid.UUID
	AssigneeUserID     uuid.UUID
	Status             domain.TicketStatus
	Kind               domain.Kind
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
