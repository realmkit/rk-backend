package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Definition configures one ticket workflow.
type Definition struct {
	ID                      uuid.UUID        `json:"id"`
	Key                     Key              `json:"key"`
	Name                    string           `json:"name"`
	Description             string           `json:"description"`
	Kind                    Kind             `json:"kind"`
	Status                  DefinitionStatus `json:"status"`
	DefaultTeamGroupID      *uuid.UUID       `json:"default_team_group_id,omitempty"`
	DefaultAssigneeUserID   *uuid.UUID       `json:"default_assignee_user_id,omitempty"`
	SubmitterCanClose       bool             `json:"submitter_can_close"`
	SubmitterCanReopen      bool             `json:"submitter_can_reopen"`
	AllowAnonymousSubmitter bool             `json:"allow_anonymous_submitter"`
	RequiresTargetUser      bool             `json:"requires_target_user"`
	RequiresPunishment      bool             `json:"requires_punishment"`
	RequiresEvidence        bool             `json:"requires_evidence"`
	MaxOpenPerSubmitter     int              `json:"max_open_per_submitter"`
	ReopenWindowSeconds     int64            `json:"reopen_window_seconds"`
	SLAFirstResponseSeconds int64            `json:"sla_first_response_seconds"`
	SLAResolutionSeconds    int64            `json:"sla_resolution_seconds"`
	MetadataSchemaKey       string           `json:"metadata_schema_key"`
	DisplayOrder            int              `json:"display_order"`
	Version                 uint64           `json:"version"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
}

// Ticket is one private staff-managed case.
type Ticket struct {
	ID                      uuid.UUID    `json:"id"`
	DefinitionID            uuid.UUID    `json:"definition_id"`
	Key                     string       `json:"key"`
	Title                   string       `json:"title"`
	Kind                    Kind         `json:"kind"`
	Status                  TicketStatus `json:"status"`
	Priority                Priority     `json:"priority"`
	SubmitterUserID         *uuid.UUID   `json:"submitter_user_id,omitempty"`
	TargetUserID            *uuid.UUID   `json:"target_user_id,omitempty"`
	PunishmentID            *uuid.UUID   `json:"punishment_id,omitempty"`
	CurrentTeamGroupID      *uuid.UUID   `json:"current_team_group_id,omitempty"`
	AssigneeUserID          *uuid.UUID   `json:"assignee_user_id,omitempty"`
	OpenedAt                time.Time    `json:"opened_at"`
	FirstStaffResponseAt    *time.Time   `json:"first_staff_response_at,omitempty"`
	LastMessageAt           *time.Time   `json:"last_message_at,omitempty"`
	LastMessageAuthorUserID *uuid.UUID   `json:"last_message_author_user_id,omitempty"`
	ClosedAt                *time.Time   `json:"closed_at,omitempty"`
	ClosedByUserID          *uuid.UUID   `json:"closed_by_user_id,omitempty"`
	CloseReason             string       `json:"close_reason,omitempty"`
	Resolution              string       `json:"resolution,omitempty"`
	EscalationLevel         int          `json:"escalation_level"`
	SLAFirstResponseDueAt   *time.Time   `json:"sla_first_response_due_at,omitempty"`
	SLAResolutionDueAt      *time.Time   `json:"sla_resolution_due_at,omitempty"`
	MessageCount            int64        `json:"message_count"`
	StaffMessageCount       int64        `json:"staff_message_count"`
	EvidenceCount           int64        `json:"evidence_count"`
	Version                 uint64       `json:"version"`
	CreatedAt               time.Time    `json:"created_at"`
	UpdatedAt               time.Time    `json:"updated_at"`
}

// Message is one private conversation entry.
type Message struct {
	ID                  uuid.UUID         `json:"id"`
	TicketID            uuid.UUID         `json:"ticket_id"`
	AuthorUserID        *uuid.UUID        `json:"author_user_id,omitempty"`
	AuthorRole          AuthorRole        `json:"author_role"`
	Visibility          MessageVisibility `json:"visibility"`
	Sequence            int64             `json:"sequence"`
	ContentFormat       string            `json:"content_format"`
	ContentDocumentJSON json.RawMessage   `json:"content_document_json"`
	ContentText         string            `json:"content_text"`
	ContentChecksum     string            `json:"content_checksum,omitempty"`
	EditCount           int64             `json:"edit_count"`
	Version             uint64            `json:"version"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// Evidence links an asset or external URL to a ticket.
type Evidence struct {
	ID                uuid.UUID         `json:"id"`
	TicketID          uuid.UUID         `json:"ticket_id"`
	MessageID         *uuid.UUID        `json:"message_id,omitempty"`
	AssetID           *uuid.UUID        `json:"asset_id,omitempty"`
	ExternalURL       string            `json:"external_url,omitempty"`
	Label             string            `json:"label"`
	Description       string            `json:"description"`
	Visibility        MessageVisibility `json:"visibility"`
	SubmittedByUserID *uuid.UUID        `json:"submitted_by_user_id,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
}

// Action records a meaningful ticket workflow effect.
type Action struct {
	ID             uuid.UUID       `json:"id"`
	TicketID       uuid.UUID       `json:"ticket_id"`
	ActorUserID    *uuid.UUID      `json:"actor_user_id,omitempty"`
	Type           ActionType      `json:"action_type"`
	Status         ActionStatus    `json:"status"`
	PayloadJSON    json.RawMessage `json:"payload_json"`
	ResultJSON     json.RawMessage `json:"result_json"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	FailedAt       *time.Time      `json:"failed_at,omitempty"`
	Error          string          `json:"error,omitempty"`
}

// DriftReport reports denormalized counter drift.
type DriftReport struct {
	Mismatches []string `json:"mismatches"`
	Repaired   bool     `json:"repaired"`
}

// Normalize returns normalized definition state.
func (definition Definition) Normalize() Definition {
	definition.Name = strings.TrimSpace(definition.Name)
	definition.Description = strings.TrimSpace(definition.Description)
	definition.MetadataSchemaKey = strings.TrimSpace(definition.MetadataSchemaKey)
	if definition.Status == "" {
		definition.Status = DefinitionActive
	}
	if definition.Version == 0 {
		definition.Version = 1
	}
	return definition
}

// Normalize returns normalized ticket state.
func (ticket Ticket) Normalize() Ticket {
	ticket.Key = strings.TrimSpace(ticket.Key)
	ticket.Title = strings.TrimSpace(ticket.Title)
	ticket.CloseReason = strings.TrimSpace(ticket.CloseReason)
	ticket.Resolution = strings.TrimSpace(ticket.Resolution)
	if ticket.Status == "" {
		ticket.Status = StatusOpen
	}
	if ticket.Priority == "" {
		ticket.Priority = PriorityNormal
	}
	if ticket.Version == 0 {
		ticket.Version = 1
	}
	return ticket
}
