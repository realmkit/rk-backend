// Package postgres stores tickets in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// DefinitionModel is the GORM model for ticket definitions.
type DefinitionModel struct {
	orm.ID
	Key                     string `gorm:"size:64;not null;uniqueIndex"`
	Name                    string `gorm:"size:120;not null"`
	Description             string `gorm:"size:1000;not null;default:''"`
	Kind                    string `gorm:"size:64;not null;index"`
	Status                  string `gorm:"size:64;not null;index"`
	DefaultTeamGroupID      *uuid.UUID
	DefaultAssigneeUserID   *uuid.UUID
	SubmitterCanClose       bool
	SubmitterCanReopen      bool
	AllowAnonymousSubmitter bool
	RequiresTargetUser      bool
	RequiresPunishment      bool
	RequiresEvidence        bool
	MaxOpenPerSubmitter     int
	ReopenWindowSeconds     int64
	SLAFirstResponseSeconds int64
	SLAResolutionSeconds    int64
	MetadataSchemaKey       string `gorm:"size:64;not null;default:''"`
	DisplayOrder            int    `gorm:"not null;default:0;index"`
	Version                 uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (DefinitionModel) TableName() string { return "ticket_definitions" }

// TicketModel is the GORM model for tickets.
type TicketModel struct {
	orm.ID
	DefinitionID            uuid.UUID `gorm:"type:uuid;not null;index"`
	Key                     string    `gorm:"size:200;uniqueIndex"`
	Title                   string    `gorm:"size:180;not null"`
	Kind                    string    `gorm:"size:64;not null;index"`
	Status                  string    `gorm:"size:64;not null;index"`
	Priority                string    `gorm:"size:64;not null;index"`
	SubmitterUserID         *uuid.UUID
	TargetUserID            *uuid.UUID
	PunishmentID            *uuid.UUID
	CurrentTeamGroupID      *uuid.UUID
	AssigneeUserID          *uuid.UUID
	OpenedAt                time.Time `gorm:"not null;index"`
	FirstStaffResponseAt    *time.Time
	LastMessageAt           *time.Time
	LastMessageAuthorUserID *uuid.UUID
	ClosedAt                *time.Time
	ClosedByUserID          *uuid.UUID
	CloseReason             string `gorm:"size:1000;not null;default:''"`
	Resolution              string `gorm:"size:1000;not null;default:''"`
	EscalationLevel         int
	SLAFirstResponseDueAt   *time.Time
	SLAResolutionDueAt      *time.Time
	MessageCount            int64
	StaffMessageCount       int64
	EvidenceCount           int64
	Version                 uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (TicketModel) TableName() string { return "tickets" }

// MessageModel is the GORM model for ticket messages.
type MessageModel struct {
	orm.ID
	TicketID            uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorUserID        *uuid.UUID
	AuthorRole          string `gorm:"size:64;not null;index"`
	Visibility          string `gorm:"size:64;not null;index"`
	Sequence            int64  `gorm:"not null;index"`
	ContentFormat       string `gorm:"size:64;not null"`
	ContentDocumentJSON string `gorm:"type:jsonb;not null"`
	ContentText         string `gorm:"type:text;not null"`
	ContentChecksum     string `gorm:"size:128;not null;default:''"`
	EditCount           int64
	Version             uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (MessageModel) TableName() string { return "ticket_messages" }

// EvidenceModel is the GORM model for ticket evidence.
type EvidenceModel struct {
	orm.ID
	TicketID          uuid.UUID `gorm:"type:uuid;not null;index"`
	MessageID         *uuid.UUID
	AssetID           *uuid.UUID
	ExternalURL       string `gorm:"size:2048;not null;default:''"`
	Label             string `gorm:"size:160;not null;default:''"`
	Description       string `gorm:"size:1000;not null;default:''"`
	Visibility        string `gorm:"size:64;not null;index"`
	SubmittedByUserID *uuid.UUID
	CreatedAt         time.Time `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the database table name.
func (EvidenceModel) TableName() string { return "ticket_evidence" }

// ActionModel is the GORM model for ticket workflow actions.
type ActionModel struct {
	orm.ID
	TicketID       uuid.UUID `gorm:"type:uuid;not null;index"`
	ActorUserID    *uuid.UUID
	Type           string `gorm:"size:64;not null;index"`
	Status         string `gorm:"size:64;not null;index"`
	PayloadJSON    string `gorm:"type:jsonb;not null;default:'{}'"`
	ResultJSON     string `gorm:"type:jsonb;not null;default:'{}'"`
	IdempotencyKey string `gorm:"size:200;uniqueIndex"`
	CreatedAt      time.Time
	CompletedAt    *time.Time
	FailedAt       *time.Time
	Error          string `gorm:"size:1000;not null;default:''"`
}

// TableName returns the database table name.
func (ActionModel) TableName() string { return "ticket_actions" }
