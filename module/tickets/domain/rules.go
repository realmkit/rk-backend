package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validate validates definition configuration.
func (definition Definition) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateKey("key", definition.Key)...)
	violations = append(violations, ValidateKind("kind", definition.Kind)...)
	violations = append(violations, ValidateDefinitionStatus("status", definition.Status)...)
	if definition.Name == "" {
		violations = AddViolation(violations, "name", "is required")
	}
	if definition.Kind == KindAppeal && !definition.RequiresPunishment {
		violations = AddViolation(violations, "requires_punishment", "appeals require a punishment")
	}
	if definition.Kind == KindReport && !definition.RequiresTargetUser {
		violations = AddViolation(violations, "requires_target_user", "reports require a target user")
	}
	if definition.MaxOpenPerSubmitter < 0 {
		violations = AddViolation(violations, "max_open_per_submitter", "must be nonnegative")
	}
	return NewValidationError(violations)
}

// Validate validates ticket state.
func (ticket Ticket) Validate() error {
	var violations []Violation
	if ticket.DefinitionID == uuid.Nil {
		violations = AddViolation(violations, "definition_id", "is required")
	}
	violations = append(violations, ValidateKind("kind", ticket.Kind)...)
	violations = append(violations, ValidateTicketStatus("status", ticket.Status)...)
	if ticket.Title == "" {
		violations = AddViolation(violations, "title", "is required")
	}
	if ticket.OpenedAt.IsZero() {
		violations = AddViolation(violations, "opened_at", "is required")
	}
	return NewValidationError(violations)
}

// Validate validates message state.
func (message Message) Validate() error {
	var violations []Violation
	if message.TicketID == uuid.Nil {
		violations = AddViolation(violations, "ticket_id", "is required")
	}
	if message.Sequence < 0 {
		violations = AddViolation(violations, "sequence", "must be nonnegative")
	}
	if !json.Valid(message.ContentDocumentJSON) || len(message.ContentDocumentJSON) == 0 {
		violations = AddViolation(violations, "content_document_json", "must be valid JSON")
	}
	if strings.TrimSpace(message.ContentText) == "" && message.AuthorRole != RoleSystem {
		violations = AddViolation(violations, "content_text", "is required")
	}
	if !validVisibility(message.Visibility) {
		violations = AddViolation(violations, "visibility", "is not supported")
	}
	return NewValidationError(violations)
}

// Validate validates evidence state.
func (evidence Evidence) Validate() error {
	var violations []Violation
	if evidence.TicketID == uuid.Nil {
		violations = AddViolation(violations, "ticket_id", "is required")
	}
	if evidence.AssetID == nil && strings.TrimSpace(evidence.ExternalURL) == "" {
		violations = AddViolation(violations, "evidence", "requires asset or external URL")
	}
	if !validVisibility(evidence.Visibility) {
		violations = AddViolation(violations, "visibility", "is not supported")
	}
	return NewValidationError(violations)
}

// Validate validates action state.
func (action Action) Validate() error {
	var violations []Violation
	if action.TicketID == uuid.Nil {
		violations = AddViolation(violations, "ticket_id", "is required")
	}
	if !validAction(action.Type) {
		violations = AddViolation(violations, "action_type", "is not supported")
	}
	if !validActionStatus(action.Status) {
		violations = AddViolation(violations, "status", "is not supported")
	}
	if len(action.PayloadJSON) > 0 && !json.Valid(action.PayloadJSON) {
		violations = AddViolation(violations, "payload_json", "must be valid JSON")
	}
	return NewValidationError(violations)
}

// CanTransition reports whether status can move to next.
func CanTransition(current TicketStatus, next TicketStatus) bool {
	if current == next {
		return true
	}
	if slicesClosed(current) {
		return next == StatusOpen
	}
	switch current {
	case StatusOpen, StatusPendingStaff, StatusPendingSubmitter, StatusEscalated:
		return true
	default:
		return next == StatusClosed
	}
}

// SLADueAt returns first response and resolution due times.
func SLADueAt(openedAt time.Time, definition Definition) (*time.Time, *time.Time) {
	var first *time.Time
	var resolution *time.Time
	if definition.SLAFirstResponseSeconds > 0 {
		value := openedAt.Add(time.Duration(definition.SLAFirstResponseSeconds) * time.Second)
		first = &value
	}
	if definition.SLAResolutionSeconds > 0 {
		value := openedAt.Add(time.Duration(definition.SLAResolutionSeconds) * time.Second)
		resolution = &value
	}
	return first, resolution
}

// MessageVisibleToSubmitter reports whether the submitter may read message.
func MessageVisibleToSubmitter(message Message) bool {
	return message.Visibility == VisibilityParticipants
}

// validVisibility reports whether visibility is supported.
func validVisibility(visibility MessageVisibility) bool {
	return visibility == VisibilityParticipants ||
		visibility == VisibilityStaffOnly ||
		visibility == VisibilitySystemOnly
}

// validAction reports whether action type is supported.
func validAction(action ActionType) bool {
	switch action {
	case ActionAssign, ActionEscalate, ActionClose, ActionReopen,
		ActionAcceptAppeal, ActionRejectAppeal, ActionRevokePunishment,
		ActionIssuePunishment:
		return true
	default:
		return false
	}
}

// validActionStatus reports whether action status is supported.
func validActionStatus(status ActionStatus) bool {
	return status == ActionPending || status == ActionCompleted || status == ActionFailed
}

// slicesClosed reports whether a status is terminal.
func slicesClosed(status TicketStatus) bool {
	return status == StatusAccepted || status == StatusRejected ||
		status == StatusResolved || status == StatusClosed ||
		status == StatusCancelled || status == StatusSpam
}
