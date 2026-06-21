// Package domain owns ticket entities and pure workflow rules.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

// Key is a stable lower snake case identifier.
type Key string

// Kind identifies a ticket workflow family.
type Kind string

// DefinitionStatus is the lifecycle of a ticket definition.
type DefinitionStatus string

// TicketStatus is the lifecycle of a ticket.
type TicketStatus string

// Priority controls staff queue ordering.
type Priority string

// AuthorRole identifies the message author role.
type AuthorRole string

// MessageVisibility controls message audience.
type MessageVisibility string

// ActionType identifies one auditable ticket action.
type ActionType string

// ActionStatus is the lifecycle of a ticket action.
type ActionStatus string

const (
	// KindAppeal is a punishment appeal.
	KindAppeal Kind = "appeal"
	// KindReport is a player report.
	KindReport Kind = "report"
	// KindSupport is a support request.
	KindSupport Kind = "support"
	// KindApplication is a staff or community application.
	KindApplication Kind = "application"
	// KindBug is a bug report.
	KindBug Kind = "bug"
	// KindPayment is a payment support ticket.
	KindPayment Kind = "payment"
	// KindCustom is an instance-specific workflow.
	KindCustom Kind = "custom"
)

const (
	// DefinitionActive accepts new tickets.
	DefinitionActive DefinitionStatus = "active"
	// DefinitionDisabled is hidden from intake.
	DefinitionDisabled DefinitionStatus = "disabled"
	// DefinitionArchived is retained for historical reads.
	DefinitionArchived DefinitionStatus = "archived"
)

const (
	// StatusOpen means the ticket is active.
	StatusOpen TicketStatus = "open"
	// StatusPendingSubmitter waits for submitter input.
	StatusPendingSubmitter TicketStatus = "pending_submitter"
	// StatusPendingStaff waits for staff input.
	StatusPendingStaff TicketStatus = "pending_staff"
	// StatusEscalated has moved to a higher review team.
	StatusEscalated TicketStatus = "escalated"
	// StatusResolved is resolved without appeal acceptance semantics.
	StatusResolved TicketStatus = "resolved"
	// StatusRejected is rejected by staff.
	StatusRejected TicketStatus = "rejected"
	// StatusAccepted is accepted by staff.
	StatusAccepted TicketStatus = "accepted"
	// StatusClosed is closed with no further replies.
	StatusClosed TicketStatus = "closed"
	// StatusCancelled is cancelled by the submitter or staff.
	StatusCancelled TicketStatus = "cancelled"
	// StatusSpam is closed as spam or abuse.
	StatusSpam TicketStatus = "spam"
)

const (
	// PriorityLow is low priority.
	PriorityLow Priority = "low"
	// PriorityNormal is normal priority.
	PriorityNormal Priority = "normal"
	// PriorityHigh is high priority.
	PriorityHigh Priority = "high"
	// PriorityUrgent is urgent priority.
	PriorityUrgent Priority = "urgent"
)

const (
	// RoleSubmitter is the ticket author.
	RoleSubmitter AuthorRole = "submitter"
	// RoleStaff is a staff participant.
	RoleStaff AuthorRole = "staff"
	// RoleSystem is an automatic message.
	RoleSystem AuthorRole = "system"
	// RoleIntegration is an external integration.
	RoleIntegration AuthorRole = "integration"
)

const (
	// VisibilityParticipants is visible to submitter and staff.
	VisibilityParticipants MessageVisibility = "public_to_participants"
	// VisibilityStaffOnly is visible to staff only.
	VisibilityStaffOnly MessageVisibility = "staff_only"
	// VisibilitySystemOnly is visible to backend/system consumers only.
	VisibilitySystemOnly MessageVisibility = "system_only"
)

const (
	// ActionAssign assigns a ticket.
	ActionAssign ActionType = "assign"
	// ActionEscalate escalates a ticket.
	ActionEscalate ActionType = "escalate"
	// ActionClose closes a ticket.
	ActionClose ActionType = "close"
	// ActionReopen reopens a ticket.
	ActionReopen ActionType = "reopen"
	// ActionAcceptAppeal accepts an appeal.
	ActionAcceptAppeal ActionType = "accept_appeal"
	// ActionRejectAppeal rejects an appeal.
	ActionRejectAppeal ActionType = "reject_appeal"
	// ActionRevokePunishment revokes a punishment from an accepted appeal.
	ActionRevokePunishment ActionType = "revoke_punishment"
	// ActionIssuePunishment issues a punishment from a report.
	ActionIssuePunishment ActionType = "issue_punishment"
)

const (
	// ActionPending is waiting to execute.
	ActionPending ActionStatus = "pending"
	// ActionCompleted completed successfully.
	ActionCompleted ActionStatus = "completed"
	// ActionFailed failed and may be inspected.
	ActionFailed ActionStatus = "failed"
)

// keyPattern stores package state.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// ErrValidation reports invalid ticket state.
var ErrValidation = errors.New("ticket validation failed")

// Violation describes one invalid field.
type Violation struct {
	Field   string `json:"field"`   // Field stores the field value.
	Message string `json:"message"` // Message stores the message value.
}

// ValidationError contains validation violations.
type ValidationError struct {
	Violations []Violation `json:"violations"` // Violations stores the violations value.
}

// Error returns the validation error message.
func (err ValidationError) Error() string { return ErrValidation.Error() }

// NewValidationError returns nil when violations are empty.
func NewValidationError(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return ValidationError{Violations: violations}
}

// AddViolation appends a validation violation.
func AddViolation(violations []Violation, field string, message string) []Violation {
	return append(violations, Violation{Field: field, Message: message})
}

// ValidateKey validates a stable ticket key.
func ValidateKey(field string, key Key) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(string(key))) {
		return []Violation{{Field: field, Message: "must be lower snake case"}}
	}
	return nil
}

// ValidateKind validates a ticket definition kind.
func ValidateKind(field string, kind Kind) []Violation {
	if slices.Contains([]Kind{KindAppeal, KindReport, KindSupport, KindApplication, KindBug, KindPayment, KindCustom}, kind) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateDefinitionStatus validates a definition status.
func ValidateDefinitionStatus(field string, status DefinitionStatus) []Violation {
	if slices.Contains([]DefinitionStatus{DefinitionActive, DefinitionDisabled, DefinitionArchived}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateTicketStatus validates ticket status.
func ValidateTicketStatus(field string, status TicketStatus) []Violation {
	if slices.Contains(validStatuses(), status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// validStatuses returns supported ticket statuses.
func validStatuses() []TicketStatus {
	return []TicketStatus{
		StatusOpen, StatusPendingSubmitter, StatusPendingStaff, StatusEscalated,
		StatusResolved, StatusRejected, StatusAccepted, StatusClosed,
		StatusCancelled, StatusSpam,
	}
}
