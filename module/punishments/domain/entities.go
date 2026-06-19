package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Definition is an admin-configured punishment type.
type Definition struct {
	ID                     uuid.UUID        `json:"id"`
	Key                    Key              `json:"key"`
	Name                   string           `json:"name"`
	Description            string           `json:"description"`
	Color                  Color            `json:"color"`
	Severity               int              `json:"severity"`
	Status                 DefinitionStatus `json:"status"`
	DefaultDurationSeconds *int64           `json:"default_duration_seconds,omitempty"`
	MinDurationSeconds     *int64           `json:"min_duration_seconds,omitempty"`
	MaxDurationSeconds     *int64           `json:"max_duration_seconds,omitempty"`
	AllowPermanent         bool             `json:"allow_permanent"`
	RequiresReason         bool             `json:"requires_reason"`
	RequiresTargetIP       bool             `json:"requires_target_ip"`
	DisplayOrder           int              `json:"display_order"`
	Version                uint64           `json:"version"`
	Actions                []ActionTemplate `json:"actions,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

// ActionTemplate describes one consequence of a definition.
type ActionTemplate struct {
	ID                uuid.UUID        `json:"id"`
	DefinitionID      uuid.UUID        `json:"definition_id"`
	TargetSystem      TargetSystem     `json:"target_system"`
	ActionType        ActionType       `json:"action_type"`
	ConfigurationJSON json.RawMessage  `json:"configuration_json"`
	DisplayOrder      int              `json:"display_order"`
	Status            DefinitionStatus `json:"status"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// Punishment is one issued moderation case.
type Punishment struct {
	ID                 uuid.UUID        `json:"id"`
	DefinitionID       uuid.UUID        `json:"definition_id"`
	TargetUserID       uuid.UUID        `json:"target_user_id"`
	TargetIPHash       string           `json:"target_ip_hash,omitempty"`
	TargetIPCiphertext string           `json:"-"`
	IssuerType         IssuerType       `json:"issuer_type"`
	IssuerUserID       *uuid.UUID       `json:"issuer_user_id,omitempty"`
	IssuerKey          string           `json:"issuer_key,omitempty"`
	Reason             string           `json:"reason"`
	PrivateReason      string           `json:"private_reason,omitempty"`
	Status             PunishmentStatus `json:"status"`
	StartsAt           time.Time        `json:"starts_at"`
	ExpiresAt          *time.Time       `json:"expires_at,omitempty"`
	RevokedAt          *time.Time       `json:"revoked_at,omitempty"`
	RevokedByUserID    *uuid.UUID       `json:"revoked_by_user_id,omitempty"`
	RevocationReason   string           `json:"revocation_reason,omitempty"`
	Source             string           `json:"source,omitempty"`
	IdempotencyKey     string           `json:"idempotency_key,omitempty"`
	Version            uint64           `json:"version"`
	Snapshots          []ActionSnapshot `json:"actions,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// ActionSnapshot preserves one issued action.
type ActionSnapshot struct {
	ID                 uuid.UUID        `json:"id"`
	PunishmentID       uuid.UUID        `json:"punishment_id"`
	DefinitionActionID uuid.UUID        `json:"definition_action_id"`
	TargetSystem       TargetSystem     `json:"target_system"`
	ActionType         ActionType       `json:"action_type"`
	ConfigurationJSON  json.RawMessage  `json:"configuration_json"`
	Status             DefinitionStatus `json:"status"`
	CreatedAt          time.Time        `json:"created_at"`
}

// ActiveRestriction is the fast restriction projection.
type ActiveRestriction struct {
	ID           uuid.UUID  `json:"id"`
	PunishmentID uuid.UUID  `json:"punishment_id"`
	TargetUserID uuid.UUID  `json:"target_user_id"`
	ActionKey    string     `json:"action_key"`
	StartsAt     time.Time  `json:"starts_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// CheckResult describes whether an action is allowed.
type CheckResult struct {
	Allowed     bool               `json:"allowed"`
	Punishment  *PunishmentSummary `json:"punishment,omitempty"`
	Restriction *ActiveRestriction `json:"restriction,omitempty"`
}

// PunishmentSummary is a safe denial summary.
type PunishmentSummary struct {
	ID        uuid.UUID  `json:"id"`
	Reason    string     `json:"reason"`
	StartsAt  time.Time  `json:"starts_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CounterDrift reports one projection mismatch.
type CounterDrift struct {
	PunishmentID uuid.UUID `json:"punishment_id"`
	ActionKey    string    `json:"action_key"`
	Expected     bool      `json:"expected"`
	Actual       bool      `json:"actual"`
}

// DriftReport reports restriction projection drift.
type DriftReport struct {
	Mismatches []CounterDrift `json:"mismatches"`
	Repaired   bool           `json:"repaired"`
}

// Normalize returns a normalized definition.
func (definition Definition) Normalize() Definition {
	definition.Name = strings.TrimSpace(definition.Name)
	definition.Description = strings.TrimSpace(definition.Description)
	if definition.Status == "" {
		definition.Status = DefinitionActive
	}
	if definition.Color == "" {
		definition.Color = "#ff5555"
	}
	if definition.Version == 0 {
		definition.Version = 1
	}
	return definition
}

// Validate validates a definition.
func (definition Definition) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateKey("key", definition.Key)...)
	violations = append(violations, ValidateColor("color", definition.Color)...)
	violations = append(violations, ValidateDefinitionStatus("status", definition.Status)...)
	if strings.TrimSpace(definition.Name) == "" {
		violations = AddViolation(violations, "name", "is required")
	}
	if definition.Severity < 0 {
		violations = AddViolation(violations, "severity", "must be nonnegative")
	}
	violations = append(violations, validateDurations(definition)...)
	for _, action := range definition.Actions {
		violations = append(violations, validationViolations(action.Validate())...)
	}
	return NewValidationError(violations)
}

// Normalize returns a normalized action.
func (action ActionTemplate) Normalize() ActionTemplate {
	if action.ID == uuid.Nil {
		action.ID = uuid.New()
	}
	if action.Status == "" {
		action.Status = DefinitionActive
	}
	if len(action.ConfigurationJSON) == 0 {
		action.ConfigurationJSON = json.RawMessage(`{}`)
	}
	return action
}

// Validate validates an action template.
func (action ActionTemplate) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateTargetSystem("target_system", action.TargetSystem)...)
	violations = append(violations, ValidateActionType("action_type", action.TargetSystem, action.ActionType)...)
	violations = append(violations, ValidateDefinitionStatus("status", action.Status)...)
	if !json.Valid(action.ConfigurationJSON) {
		violations = AddViolation(violations, "configuration_json", "must be valid JSON")
	}
	return NewValidationError(violations)
}
