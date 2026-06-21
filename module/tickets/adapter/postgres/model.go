// Package postgres stores tickets in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// DefinitionModel is the GORM model for ticket definitions.
type DefinitionModel struct {
	orm.ID                             // ID embeds shared fields.
	Key                     string     `gorm:"size:64;not null;uniqueIndex"`  // Key stores the key value.
	Name                    string     `gorm:"size:120;not null"`             // Name stores the name value.
	Description             string     `gorm:"size:1000;not null;default:''"` // Description stores the description value.
	Kind                    string     `gorm:"size:64;not null;index"`        // Kind stores the kind value.
	Status                  string     `gorm:"size:64;not null;index"`        // Status stores the status value.
	DefaultTeamGroupID      *uuid.UUID // DefaultTeamGroupID stores the default team group i d value.
	DefaultAssigneeUserID   *uuid.UUID // DefaultAssigneeUserID stores the default assignee user i d value.
	SubmitterCanClose       bool       // SubmitterCanClose stores the submitter can close value.
	SubmitterCanReopen      bool       // SubmitterCanReopen stores the submitter can reopen value.
	AllowAnonymousSubmitter bool       // AllowAnonymousSubmitter stores the allow anonymous submitter value.
	RequiresTargetUser      bool       // RequiresTargetUser stores the requires target user value.
	RequiresPunishment      bool       // RequiresPunishment stores the requires punishment value.
	RequiresEvidence        bool       // RequiresEvidence stores the requires evidence value.
	MaxOpenPerSubmitter     int        // MaxOpenPerSubmitter stores the max open per submitter value.
	ReopenWindowSeconds     int64      // ReopenWindowSeconds stores the reopen window seconds value.
	SLAFirstResponseSeconds int64      // SLAFirstResponseSeconds stores the s l a first response seconds value.
	SLAResolutionSeconds    int64      // SLAResolutionSeconds stores the s l a resolution seconds value.
	MetadataSchemaKey       string     `gorm:"size:64;not null;default:''"` // MetadataSchemaKey stores the metadata schema key value.
	DisplayOrder            int        `gorm:"not null;default:0;index"`    // DisplayOrder stores the display order value.
	Version                 uint64     `gorm:"not null;default:1"`          // Version stores the version value.
	orm.Timestamps                     // Timestamps embeds shared fields.
	orm.SoftDelete                     // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (DefinitionModel) TableName() string { return "ticket_definitions" }

// TicketModel is the GORM model for tickets.
type TicketModel struct {
	orm.ID                             // ID embeds shared fields.
	DefinitionID            uuid.UUID  `gorm:"type:uuid;not null;index"` // DefinitionID stores the definition i d value.
	Key                     string     `gorm:"size:200;uniqueIndex"`     // Key stores the key value.
	Title                   string     `gorm:"size:180;not null"`        // Title stores the title value.
	Kind                    string     `gorm:"size:64;not null;index"`   // Kind stores the kind value.
	Status                  string     `gorm:"size:64;not null;index"`   // Status stores the status value.
	Priority                string     `gorm:"size:64;not null;index"`   // Priority stores the priority value.
	SubmitterUserID         *uuid.UUID // SubmitterUserID stores the submitter user i d value.
	TargetUserID            *uuid.UUID // TargetUserID stores the target user i d value.
	PunishmentID            *uuid.UUID // PunishmentID stores the punishment i d value.
	CurrentTeamGroupID      *uuid.UUID // CurrentTeamGroupID stores the current team group i d value.
	AssigneeUserID          *uuid.UUID // AssigneeUserID stores the assignee user i d value.
	OpenedAt                time.Time  `gorm:"not null;index"` // OpenedAt stores the opened at value.
	FirstStaffResponseAt    *time.Time // FirstStaffResponseAt stores the first staff response at value.
	LastMessageAt           *time.Time // LastMessageAt stores the last message at value.
	LastMessageAuthorUserID *uuid.UUID // LastMessageAuthorUserID stores the last message author user i d value.
	ClosedAt                *time.Time // ClosedAt stores the closed at value.
	ClosedByUserID          *uuid.UUID // ClosedByUserID stores the closed by user i d value.
	CloseReason             string     `gorm:"size:1000;not null;default:''"` // CloseReason stores the close reason value.
	Resolution              string     `gorm:"size:1000;not null;default:''"` // Resolution stores the resolution value.
	EscalationLevel         int        // EscalationLevel stores the escalation level value.
	SLAFirstResponseDueAt   *time.Time // SLAFirstResponseDueAt stores the s l a first response due at value.
	SLAResolutionDueAt      *time.Time // SLAResolutionDueAt stores the s l a resolution due at value.
	MessageCount            int64      // MessageCount stores the message count value.
	StaffMessageCount       int64      // StaffMessageCount stores the staff message count value.
	EvidenceCount           int64      // EvidenceCount stores the evidence count value.
	Version                 uint64     `gorm:"not null;default:1"` // Version stores the version value.
	orm.Timestamps                     // Timestamps embeds shared fields.
	orm.SoftDelete                     // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (TicketModel) TableName() string { return "tickets" }

// MessageModel is the GORM model for ticket messages.
type MessageModel struct {
	orm.ID                         // ID embeds shared fields.
	TicketID            uuid.UUID  `gorm:"type:uuid;not null;index"` // TicketID stores the ticket i d value.
	AuthorUserID        *uuid.UUID // AuthorUserID stores the author user i d value.
	AuthorRole          string     `gorm:"size:64;not null;index"`       // AuthorRole stores the author role value.
	Visibility          string     `gorm:"size:64;not null;index"`       // Visibility stores the visibility value.
	Sequence            int64      `gorm:"not null;index"`               // Sequence stores the sequence value.
	ContentFormat       string     `gorm:"size:64;not null"`             // ContentFormat stores the content format value.
	ContentDocumentJSON string     `gorm:"type:jsonb;not null"`          // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string     `gorm:"type:text;not null"`           // ContentText stores the content text value.
	ContentChecksum     string     `gorm:"size:128;not null;default:''"` // ContentChecksum stores the content checksum value.
	EditCount           int64      // EditCount stores the edit count value.
	Version             uint64     `gorm:"not null;default:1"` // Version stores the version value.
	orm.Timestamps                 // Timestamps embeds shared fields.
	orm.SoftDelete                 // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (MessageModel) TableName() string { return "ticket_messages" }

// EvidenceModel is the GORM model for ticket evidence.
type EvidenceModel struct {
	orm.ID                       // ID embeds shared fields.
	TicketID          uuid.UUID  `gorm:"type:uuid;not null;index"` // TicketID stores the ticket i d value.
	MessageID         *uuid.UUID // MessageID stores the message i d value.
	AssetID           *uuid.UUID // AssetID stores the asset i d value.
	ExternalURL       string     `gorm:"size:2048;not null;default:''"` // ExternalURL stores the external u r l value.
	Label             string     `gorm:"size:160;not null;default:''"`  // Label stores the label value.
	Description       string     `gorm:"size:1000;not null;default:''"` // Description stores the description value.
	Visibility        string     `gorm:"size:64;not null;index"`        // Visibility stores the visibility value.
	SubmittedByUserID *uuid.UUID // SubmittedByUserID stores the submitted by user i d value.
	CreatedAt         time.Time  `gorm:"not null"` // CreatedAt stores the created at value.
	orm.SoftDelete               // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (EvidenceModel) TableName() string { return "ticket_evidence" }

// ActionModel is the GORM model for ticket workflow actions.
type ActionModel struct {
	orm.ID                    // ID embeds shared fields.
	TicketID       uuid.UUID  `gorm:"type:uuid;not null;index"` // TicketID stores the ticket i d value.
	ActorUserID    *uuid.UUID // ActorUserID stores the actor user i d value.
	Type           string     `gorm:"size:64;not null;index"`           // Type stores the type value.
	Status         string     `gorm:"size:64;not null;index"`           // Status stores the status value.
	PayloadJSON    string     `gorm:"type:jsonb;not null;default:'{}'"` // PayloadJSON stores the payload j s o n value.
	ResultJSON     string     `gorm:"type:jsonb;not null;default:'{}'"` // ResultJSON stores the result j s o n value.
	IdempotencyKey string     `gorm:"size:200;uniqueIndex"`             // IdempotencyKey stores the idempotency key value.
	CreatedAt      time.Time  // CreatedAt stores the created at value.
	CompletedAt    *time.Time // CompletedAt stores the completed at value.
	FailedAt       *time.Time // FailedAt stores the failed at value.
	Error          string     `gorm:"size:1000;not null;default:''"` // Error stores the error value.
}

// TableName returns the database table name.
func (ActionModel) TableName() string { return "ticket_actions" }
