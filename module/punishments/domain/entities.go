package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Definition is an admin-configured punishment type.
type Definition struct {
	ID                     uuid.UUID        `json:"id"`                                 // ID stores the i d value.
	Key                    Key              `json:"key"`                                // Key stores the key value.
	Name                   string           `json:"name"`                               // Name stores the name value.
	Description            string           `json:"description"`                        // Description stores the description value.
	Color                  Color            `json:"color"`                              // Color stores the color value.
	Severity               int              `json:"severity"`                           // Severity stores the severity value.
	Status                 DefinitionStatus `json:"status"`                             // Status stores the status value.
	DefaultDurationSeconds *int64           `json:"default_duration_seconds,omitempty"` // DefaultDurationSeconds stores the default duration seconds value.
	MinDurationSeconds     *int64           `json:"min_duration_seconds,omitempty"`     // MinDurationSeconds stores the min duration seconds value.
	MaxDurationSeconds     *int64           `json:"max_duration_seconds,omitempty"`     // MaxDurationSeconds stores the max duration seconds value.
	AllowPermanent         bool             `json:"allow_permanent"`                    // AllowPermanent stores the allow permanent value.
	RequiresReason         bool             `json:"requires_reason"`                    // RequiresReason stores the requires reason value.
	RequiresTargetIP       bool             `json:"requires_target_ip"`                 // RequiresTargetIP stores the requires target i p value.
	DisplayOrder           int              `json:"display_order"`                      // DisplayOrder stores the display order value.
	Version                uint64           `json:"version"`                            // Version stores the version value.
	Actions                []ActionTemplate `json:"actions,omitempty"`                  // Actions stores the actions value.
	CreatedAt              time.Time        `json:"created_at"`                         // CreatedAt stores the created at value.
	UpdatedAt              time.Time        `json:"updated_at"`                         // UpdatedAt stores the updated at value.
}

// ActionTemplate describes one consequence of a definition.
type ActionTemplate struct {
	ID                uuid.UUID        `json:"id"`                 // ID stores the i d value.
	DefinitionID      uuid.UUID        `json:"definition_id"`      // DefinitionID stores the definition i d value.
	TargetSystem      TargetSystem     `json:"target_system"`      // TargetSystem stores the target system value.
	ActionType        ActionType       `json:"action_type"`        // ActionType stores the action type value.
	ConfigurationJSON json.RawMessage  `json:"configuration_json"` // ConfigurationJSON stores the configuration j s o n value.
	DisplayOrder      int              `json:"display_order"`      // DisplayOrder stores the display order value.
	Status            DefinitionStatus `json:"status"`             // Status stores the status value.
	CreatedAt         time.Time        `json:"created_at"`         // CreatedAt stores the created at value.
	UpdatedAt         time.Time        `json:"updated_at"`         // UpdatedAt stores the updated at value.
}

// Punishment is one issued moderation case.
type Punishment struct {
	ID                 uuid.UUID        `json:"id"`                           // ID stores the i d value.
	DefinitionID       uuid.UUID        `json:"definition_id"`                // DefinitionID stores the definition i d value.
	TargetUserID       uuid.UUID        `json:"target_user_id"`               // TargetUserID stores the target user i d value.
	TargetIPHash       string           `json:"target_ip_hash,omitempty"`     // TargetIPHash stores the target i p hash value.
	TargetIPCiphertext string           `json:"-"`                            // TargetIPCiphertext stores the target i p ciphertext value.
	IssuerType         IssuerType       `json:"issuer_type"`                  // IssuerType stores the issuer type value.
	IssuerUserID       *uuid.UUID       `json:"issuer_user_id,omitempty"`     // IssuerUserID stores the issuer user i d value.
	IssuerKey          string           `json:"issuer_key,omitempty"`         // IssuerKey stores the issuer key value.
	Reason             string           `json:"reason"`                       // Reason stores the reason value.
	PrivateReason      string           `json:"private_reason,omitempty"`     // PrivateReason stores the private reason value.
	Status             PunishmentStatus `json:"status"`                       // Status stores the status value.
	StartsAt           time.Time        `json:"starts_at"`                    // StartsAt stores the starts at value.
	ExpiresAt          *time.Time       `json:"expires_at,omitempty"`         // ExpiresAt stores the expires at value.
	RevokedAt          *time.Time       `json:"revoked_at,omitempty"`         // RevokedAt stores the revoked at value.
	RevokedByUserID    *uuid.UUID       `json:"revoked_by_user_id,omitempty"` // RevokedByUserID stores the revoked by user i d value.
	RevocationReason   string           `json:"revocation_reason,omitempty"`  // RevocationReason stores the revocation reason value.
	Source             string           `json:"source,omitempty"`             // Source stores the source value.
	IdempotencyKey     string           `json:"idempotency_key,omitempty"`    // IdempotencyKey stores the idempotency key value.
	Version            uint64           `json:"version"`                      // Version stores the version value.
	Snapshots          []ActionSnapshot `json:"actions,omitempty"`            // Snapshots stores the snapshots value.
	CreatedAt          time.Time        `json:"created_at"`                   // CreatedAt stores the created at value.
	UpdatedAt          time.Time        `json:"updated_at"`                   // UpdatedAt stores the updated at value.
}

// ActionSnapshot preserves one issued action.
type ActionSnapshot struct {
	ID                 uuid.UUID        `json:"id"`                   // ID stores the i d value.
	PunishmentID       uuid.UUID        `json:"punishment_id"`        // PunishmentID stores the punishment i d value.
	DefinitionActionID uuid.UUID        `json:"definition_action_id"` // DefinitionActionID stores the definition action i d value.
	TargetSystem       TargetSystem     `json:"target_system"`        // TargetSystem stores the target system value.
	ActionType         ActionType       `json:"action_type"`          // ActionType stores the action type value.
	ConfigurationJSON  json.RawMessage  `json:"configuration_json"`   // ConfigurationJSON stores the configuration j s o n value.
	Status             DefinitionStatus `json:"status"`               // Status stores the status value.
	CreatedAt          time.Time        `json:"created_at"`           // CreatedAt stores the created at value.
}

// ActiveRestriction is the fast restriction projection.
type ActiveRestriction struct {
	ID           uuid.UUID  `json:"id"`                   // ID stores the i d value.
	PunishmentID uuid.UUID  `json:"punishment_id"`        // PunishmentID stores the punishment i d value.
	TargetUserID uuid.UUID  `json:"target_user_id"`       // TargetUserID stores the target user i d value.
	ActionKey    string     `json:"action_key"`           // ActionKey stores the action key value.
	StartsAt     time.Time  `json:"starts_at"`            // StartsAt stores the starts at value.
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // ExpiresAt stores the expires at value.
	CreatedAt    time.Time  `json:"created_at"`           // CreatedAt stores the created at value.
}

// CheckResult describes whether an action is allowed.
type CheckResult struct {
	Allowed     bool               `json:"allowed"`               // Allowed stores the allowed value.
	Punishment  *PunishmentSummary `json:"punishment,omitempty"`  // Punishment stores the punishment value.
	Restriction *ActiveRestriction `json:"restriction,omitempty"` // Restriction stores the restriction value.
}

// PunishmentSummary is a safe denial summary.
type PunishmentSummary struct {
	ID        uuid.UUID  `json:"id"`                   // ID stores the i d value.
	Reason    string     `json:"reason"`               // Reason stores the reason value.
	StartsAt  time.Time  `json:"starts_at"`            // StartsAt stores the starts at value.
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // ExpiresAt stores the expires at value.
}

// CounterDrift reports one projection mismatch.
type CounterDrift struct {
	PunishmentID uuid.UUID `json:"punishment_id"` // PunishmentID stores the punishment i d value.
	ActionKey    string    `json:"action_key"`    // ActionKey stores the action key value.
	Expected     bool      `json:"expected"`      // Expected stores the expected value.
	Actual       bool      `json:"actual"`        // Actual stores the actual value.
}

// DriftReport reports restriction projection drift.
type DriftReport struct {
	Mismatches []CounterDrift `json:"mismatches"` // Mismatches stores the mismatches value.
	Repaired   bool           `json:"repaired"`   // Repaired stores the repaired value.
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
