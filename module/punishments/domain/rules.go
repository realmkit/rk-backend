package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Normalize returns a normalized punishment.
func (punishment Punishment) Normalize() Punishment {
	punishment.Reason = strings.TrimSpace(punishment.Reason)
	punishment.PrivateReason = strings.TrimSpace(punishment.PrivateReason)
	punishment.IssuerKey = strings.TrimSpace(punishment.IssuerKey)
	punishment.Source = strings.TrimSpace(punishment.Source)
	if punishment.Status == "" {
		punishment.Status = PunishmentActive
	}
	if punishment.Version == 0 {
		punishment.Version = 1
	}
	return punishment
}

// Validate validates an issued punishment.
func (punishment Punishment) Validate() error {
	var violations []Violation
	if punishment.DefinitionID == uuid.Nil {
		violations = AddViolation(violations, "definition_id", "is required")
	}
	if punishment.TargetUserID == uuid.Nil {
		violations = AddViolation(violations, "target_user_id", "is required")
	}
	violations = append(violations, ValidateIssuerType("issuer_type", punishment.IssuerType)...)
	violations = append(violations, ValidatePunishmentStatus("status", punishment.Status)...)
	if punishment.IssuerType == IssuerUser && (punishment.IssuerUserID == nil || *punishment.IssuerUserID == uuid.Nil) {
		violations = AddViolation(violations, "issuer_user_id", "is required for user issuer")
	}
	if punishment.IssuerType != IssuerUser && punishment.IssuerKey == "" {
		violations = AddViolation(violations, "issuer_key", "is required for non-user issuer")
	}
	if punishment.Reason == "" {
		violations = AddViolation(violations, "reason", "is required")
	}
	if punishment.StartsAt.IsZero() {
		violations = AddViolation(violations, "starts_at", "is required")
	}
	if punishment.ExpiresAt != nil && !punishment.ExpiresAt.After(punishment.StartsAt) {
		violations = AddViolation(violations, "expires_at", "must be after starts_at")
	}
	for _, snapshot := range punishment.Snapshots {
		violations = append(violations, validationViolations(snapshot.Validate())...)
	}
	return NewValidationError(violations)
}

// ActiveAt reports whether punishment applies at now.
func (punishment Punishment) ActiveAt(now time.Time) bool {
	if punishment.Status != PunishmentActive {
		return false
	}
	if now.Before(punishment.StartsAt) {
		return false
	}
	return punishment.ExpiresAt == nil || now.Before(*punishment.ExpiresAt)
}

// ValidateIssueDuration validates requested duration against definition rules.
func ValidateIssueDuration(definition Definition, startsAt time.Time, expiresAt *time.Time) error {
	var violations []Violation
	if expiresAt == nil {
		if !definition.AllowPermanent {
			violations = AddViolation(violations, "expires_at", "permanent punishment is not allowed")
		}
		return NewValidationError(violations)
	}
	duration := expiresAt.Sub(startsAt)
	if definition.MinDurationSeconds != nil && duration < time.Duration(*definition.MinDurationSeconds)*time.Second {
		violations = AddViolation(violations, "expires_at", "duration is shorter than minimum")
	}
	if definition.MaxDurationSeconds != nil && duration > time.Duration(*definition.MaxDurationSeconds)*time.Second {
		violations = AddViolation(violations, "expires_at", "duration is longer than maximum")
	}
	return NewValidationError(violations)
}

// SnapshotFromTemplate copies an action template into a punishment snapshot.
func SnapshotFromTemplate(punishmentID uuid.UUID, action ActionTemplate) ActionSnapshot {
	return ActionSnapshot{
		ID:                 uuid.New(),
		PunishmentID:       punishmentID,
		DefinitionActionID: action.ID,
		TargetSystem:       action.TargetSystem,
		ActionType:         action.ActionType,
		ConfigurationJSON:  cloneJSON(action.ConfigurationJSON),
		Status:             DefinitionActive,
		CreatedAt:          time.Now().UTC(),
	}
}

// RestrictionFromSnapshot creates an active restriction when applicable.
func RestrictionFromSnapshot(punishment Punishment, snapshot ActionSnapshot) (ActiveRestriction, bool) {
	if snapshot.TargetSystem != TargetRealmKit {
		return ActiveRestriction{}, false
	}
	return ActiveRestriction{
		ID:           uuid.New(),
		PunishmentID: punishment.ID,
		TargetUserID: punishment.TargetUserID,
		ActionKey:    string(snapshot.ActionType),
		StartsAt:     punishment.StartsAt,
		ExpiresAt:    punishment.ExpiresAt,
		CreatedAt:    time.Now().UTC(),
	}, true
}

// Validate validates a snapshot.
func (snapshot ActionSnapshot) Validate() error {
	action := ActionTemplate{
		TargetSystem:      snapshot.TargetSystem,
		ActionType:        snapshot.ActionType,
		ConfigurationJSON: snapshot.ConfigurationJSON,
		Status:            snapshot.Status,
	}
	return action.Normalize().Validate()
}

// ActiveAt reports whether a restriction applies at now.
func (restriction ActiveRestriction) ActiveAt(now time.Time) bool {
	if now.Before(restriction.StartsAt) {
		return false
	}
	return restriction.ExpiresAt == nil || now.Before(*restriction.ExpiresAt)
}

// validateDurations validates configured duration bounds.
func validateDurations(definition Definition) []Violation {
	var violations []Violation
	for field, value := range map[string]*int64{
		"default_duration_seconds": definition.DefaultDurationSeconds,
		"min_duration_seconds":     definition.MinDurationSeconds,
		"max_duration_seconds":     definition.MaxDurationSeconds,
	} {
		if value != nil && *value <= 0 {
			violations = AddViolation(violations, field, "must be positive")
		}
	}
	if definition.MinDurationSeconds != nil && definition.MaxDurationSeconds != nil &&
		*definition.MinDurationSeconds > *definition.MaxDurationSeconds {
		violations = AddViolation(violations, "max_duration_seconds", "must be greater than minimum")
	}
	return violations
}

// validationViolations unwraps field violations from validation errors.
func validationViolations(err error) []Violation {
	if validation, ok := err.(ValidationError); ok {
		return validation.Violations
	}
	return nil
}

// cloneJSON returns a detached JSON buffer.
func cloneJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage{}, raw...)
}
