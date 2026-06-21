package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Definition configures one ticket workflow.
type Definition struct {
	ID                      uuid.UUID        `json:"id"`                                 // ID stores the i d value.
	Key                     Key              `json:"key"`                                // Key stores the key value.
	Name                    string           `json:"name"`                               // Name stores the name value.
	Description             string           `json:"description"`                        // Description stores the description value.
	Kind                    Kind             `json:"kind"`                               // Kind stores the kind value.
	Status                  DefinitionStatus `json:"status"`                             // Status stores the status value.
	DefaultTeamGroupID      *uuid.UUID       `json:"default_team_group_id,omitempty"`    // DefaultTeamGroupID stores the default team group i d value.
	DefaultAssigneeUserID   *uuid.UUID       `json:"default_assignee_user_id,omitempty"` // DefaultAssigneeUserID stores the default assignee user i d value.
	SubmitterCanClose       bool             `json:"submitter_can_close"`                // SubmitterCanClose stores the submitter can close value.
	SubmitterCanReopen      bool             `json:"submitter_can_reopen"`               // SubmitterCanReopen stores the submitter can reopen value.
	AllowAnonymousSubmitter bool             `json:"allow_anonymous_submitter"`          // AllowAnonymousSubmitter stores the allow anonymous submitter value.
	RequiresTargetUser      bool             `json:"requires_target_user"`               // RequiresTargetUser stores the requires target user value.
	RequiresPunishment      bool             `json:"requires_punishment"`                // RequiresPunishment stores the requires punishment value.
	RequiresEvidence        bool             `json:"requires_evidence"`                  // RequiresEvidence stores the requires evidence value.
	MaxOpenPerSubmitter     int              `json:"max_open_per_submitter"`             // MaxOpenPerSubmitter stores the max open per submitter value.
	ReopenWindowSeconds     int64            `json:"reopen_window_seconds"`              // ReopenWindowSeconds stores the reopen window seconds value.
	SLAFirstResponseSeconds int64            `json:"sla_first_response_seconds"`         // SLAFirstResponseSeconds stores the s l a first response seconds value.
	SLAResolutionSeconds    int64            `json:"sla_resolution_seconds"`             // SLAResolutionSeconds stores the s l a resolution seconds value.
	MetadataSchemaKey       string           `json:"metadata_schema_key"`                // MetadataSchemaKey stores the metadata schema key value.
	DisplayOrder            int              `json:"display_order"`                      // DisplayOrder stores the display order value.
	Version                 uint64           `json:"version"`                            // Version stores the version value.
	CreatedAt               time.Time        `json:"created_at"`                         // CreatedAt stores the created at value.
	UpdatedAt               time.Time        `json:"updated_at"`                         // UpdatedAt stores the updated at value.
}

// Ticket is one private staff-managed case.
type Ticket struct {
	ID                      uuid.UUID    `json:"id"`                                    // ID stores the i d value.
	DefinitionID            uuid.UUID    `json:"definition_id"`                         // DefinitionID stores the definition i d value.
	Key                     string       `json:"key"`                                   // Key stores the key value.
	Title                   string       `json:"title"`                                 // Title stores the title value.
	Kind                    Kind         `json:"kind"`                                  // Kind stores the kind value.
	Status                  TicketStatus `json:"status"`                                // Status stores the status value.
	Priority                Priority     `json:"priority"`                              // Priority stores the priority value.
	SubmitterUserID         *uuid.UUID   `json:"submitter_user_id,omitempty"`           // SubmitterUserID stores the submitter user i d value.
	TargetUserID            *uuid.UUID   `json:"target_user_id,omitempty"`              // TargetUserID stores the target user i d value.
	PunishmentID            *uuid.UUID   `json:"punishment_id,omitempty"`               // PunishmentID stores the punishment i d value.
	CurrentTeamGroupID      *uuid.UUID   `json:"current_team_group_id,omitempty"`       // CurrentTeamGroupID stores the current team group i d value.
	AssigneeUserID          *uuid.UUID   `json:"assignee_user_id,omitempty"`            // AssigneeUserID stores the assignee user i d value.
	OpenedAt                time.Time    `json:"opened_at"`                             // OpenedAt stores the opened at value.
	FirstStaffResponseAt    *time.Time   `json:"first_staff_response_at,omitempty"`     // FirstStaffResponseAt stores the first staff response at value.
	LastMessageAt           *time.Time   `json:"last_message_at,omitempty"`             // LastMessageAt stores the last message at value.
	LastMessageAuthorUserID *uuid.UUID   `json:"last_message_author_user_id,omitempty"` // LastMessageAuthorUserID stores the last message author user i d value.
	ClosedAt                *time.Time   `json:"closed_at,omitempty"`                   // ClosedAt stores the closed at value.
	ClosedByUserID          *uuid.UUID   `json:"closed_by_user_id,omitempty"`           // ClosedByUserID stores the closed by user i d value.
	CloseReason             string       `json:"close_reason,omitempty"`                // CloseReason stores the close reason value.
	Resolution              string       `json:"resolution,omitempty"`                  // Resolution stores the resolution value.
	EscalationLevel         int          `json:"escalation_level"`                      // EscalationLevel stores the escalation level value.
	SLAFirstResponseDueAt   *time.Time   `json:"sla_first_response_due_at,omitempty"`   // SLAFirstResponseDueAt stores the s l a first response due at value.
	SLAResolutionDueAt      *time.Time   `json:"sla_resolution_due_at,omitempty"`       // SLAResolutionDueAt stores the s l a resolution due at value.
	MessageCount            int64        `json:"message_count"`                         // MessageCount stores the message count value.
	StaffMessageCount       int64        `json:"staff_message_count"`                   // StaffMessageCount stores the staff message count value.
	EvidenceCount           int64        `json:"evidence_count"`                        // EvidenceCount stores the evidence count value.
	Version                 uint64       `json:"version"`                               // Version stores the version value.
	CreatedAt               time.Time    `json:"created_at"`                            // CreatedAt stores the created at value.
	UpdatedAt               time.Time    `json:"updated_at"`                            // UpdatedAt stores the updated at value.
}

// Message is one private conversation entry.
type Message struct {
	ID                  uuid.UUID         `json:"id"`                         // ID stores the i d value.
	TicketID            uuid.UUID         `json:"ticket_id"`                  // TicketID stores the ticket i d value.
	AuthorUserID        *uuid.UUID        `json:"author_user_id,omitempty"`   // AuthorUserID stores the author user i d value.
	AuthorRole          AuthorRole        `json:"author_role"`                // AuthorRole stores the author role value.
	Visibility          MessageVisibility `json:"visibility"`                 // Visibility stores the visibility value.
	Sequence            int64             `json:"sequence"`                   // Sequence stores the sequence value.
	ContentFormat       string            `json:"content_format"`             // ContentFormat stores the content format value.
	ContentDocumentJSON json.RawMessage   `json:"content_document_json"`      // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string            `json:"content_text"`               // ContentText stores the content text value.
	ContentChecksum     string            `json:"content_checksum,omitempty"` // ContentChecksum stores the content checksum value.
	EditCount           int64             `json:"edit_count"`                 // EditCount stores the edit count value.
	Version             uint64            `json:"version"`                    // Version stores the version value.
	CreatedAt           time.Time         `json:"created_at"`                 // CreatedAt stores the created at value.
	UpdatedAt           time.Time         `json:"updated_at"`                 // UpdatedAt stores the updated at value.
}

// Evidence links an asset or external URL to a ticket.
type Evidence struct {
	ID                uuid.UUID         `json:"id"`                             // ID stores the i d value.
	TicketID          uuid.UUID         `json:"ticket_id"`                      // TicketID stores the ticket i d value.
	MessageID         *uuid.UUID        `json:"message_id,omitempty"`           // MessageID stores the message i d value.
	AssetID           *uuid.UUID        `json:"asset_id,omitempty"`             // AssetID stores the asset i d value.
	ExternalURL       string            `json:"external_url,omitempty"`         // ExternalURL stores the external u r l value.
	Label             string            `json:"label"`                          // Label stores the label value.
	Description       string            `json:"description"`                    // Description stores the description value.
	Visibility        MessageVisibility `json:"visibility"`                     // Visibility stores the visibility value.
	SubmittedByUserID *uuid.UUID        `json:"submitted_by_user_id,omitempty"` // SubmittedByUserID stores the submitted by user i d value.
	CreatedAt         time.Time         `json:"created_at"`                     // CreatedAt stores the created at value.
}

// Action records a meaningful ticket workflow effect.
type Action struct {
	ID             uuid.UUID       `json:"id"`                        // ID stores the i d value.
	TicketID       uuid.UUID       `json:"ticket_id"`                 // TicketID stores the ticket i d value.
	ActorUserID    *uuid.UUID      `json:"actor_user_id,omitempty"`   // ActorUserID stores the actor user i d value.
	Type           ActionType      `json:"action_type"`               // Type stores the type value.
	Status         ActionStatus    `json:"status"`                    // Status stores the status value.
	PayloadJSON    json.RawMessage `json:"payload_json"`              // PayloadJSON stores the payload j s o n value.
	ResultJSON     json.RawMessage `json:"result_json"`               // ResultJSON stores the result j s o n value.
	IdempotencyKey string          `json:"idempotency_key,omitempty"` // IdempotencyKey stores the idempotency key value.
	CreatedAt      time.Time       `json:"created_at"`                // CreatedAt stores the created at value.
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`    // CompletedAt stores the completed at value.
	FailedAt       *time.Time      `json:"failed_at,omitempty"`       // FailedAt stores the failed at value.
	Error          string          `json:"error,omitempty"`           // Error stores the error value.
}

// DriftReport reports denormalized counter drift.
type DriftReport struct {
	Mismatches []string `json:"mismatches"` // Mismatches stores the mismatches value.
	Repaired   bool     `json:"repaired"`   // Repaired stores the repaired value.
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
