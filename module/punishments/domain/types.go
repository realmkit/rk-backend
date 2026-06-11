// Package domain owns punishment entities and pure validation rules.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

// Key is a stable lower snake case identifier.
type Key string

// Color is a hex presentation color.
type Color string

// DefinitionStatus is the lifecycle of a punishment definition.
type DefinitionStatus string

// TargetSystem identifies where an action applies.
type TargetSystem string

// Effect identifies what an action does.
type Effect string

// IssuerType identifies who or what issued a punishment.
type IssuerType string

// PunishmentStatus is the lifecycle of an issued punishment.
type PunishmentStatus string

const (
	// DefinitionActive can be issued.
	DefinitionActive DefinitionStatus = "active"
	// DefinitionDisabled cannot be issued.
	DefinitionDisabled DefinitionStatus = "disabled"
	// DefinitionArchived is retained for history.
	DefinitionArchived DefinitionStatus = "archived"
)

const (
	// TargetGameHub applies inside GameHub.
	TargetGameHub TargetSystem = "gamehub"
	// TargetMinecraft is a Minecraft integration target.
	TargetMinecraft TargetSystem = "minecraft"
	// TargetSAMP is a SAMP integration target.
	TargetSAMP TargetSystem = "samp"
	// TargetDiscord is a Discord integration target.
	TargetDiscord TargetSystem = "discord"
	// TargetCustom is a community-defined integration target.
	TargetCustom TargetSystem = "custom"
)

const (
	// EffectRestrict denies a GameHub action while active.
	EffectRestrict Effect = "restrict"
	// EffectNotify emits a notification-oriented event.
	EffectNotify Effect = "notify"
	// EffectExternalEvent emits an integration event.
	EffectExternalEvent Effect = "external_event"
)

const (
	// IssuerUser is a local GameHub user.
	IssuerUser IssuerType = "user"
	// IssuerSystem is internal automation.
	IssuerSystem IssuerType = "system"
	// IssuerIntegration is an external integration.
	IssuerIntegration IssuerType = "integration"
	// IssuerAnticheat is an anticheat source.
	IssuerAnticheat IssuerType = "anticheat"
	// IssuerImport is migrated historical data.
	IssuerImport IssuerType = "import"
	// IssuerUnknown is retained for legacy ambiguity.
	IssuerUnknown IssuerType = "unknown"
)

const (
	// PunishmentActive currently applies.
	PunishmentActive PunishmentStatus = "active"
	// PunishmentExpired ended naturally.
	PunishmentExpired PunishmentStatus = "expired"
	// PunishmentRevoked was removed by staff.
	PunishmentRevoked PunishmentStatus = "revoked"
	// PunishmentVoided was invalidated.
	PunishmentVoided PunishmentStatus = "voided"
	// PunishmentPending is recorded but not active yet.
	PunishmentPending PunishmentStatus = "pending"
)

const (
	// ActionForumsCreateThread denies forum thread creation.
	ActionForumsCreateThread = "gamehub.forums.create_thread"
	// ActionForumsReply denies forum replies.
	ActionForumsReply = "gamehub.forums.reply"
	// ActionForumsLikePosts denies forum likes.
	ActionForumsLikePosts = "gamehub.forums.like_posts"
	// ActionMessagesSend denies sending messages.
	ActionMessagesSend = "gamehub.messages.send"
	// ActionAssetsUpload denies asset uploads.
	ActionAssetsUpload = "gamehub.assets.upload"
	// ActionProfileUpdate denies profile updates.
	ActionProfileUpdate = "gamehub.profile.update"
)

var (
	keyPattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)
	colorPattern  = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	actionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
)

// ErrValidation reports invalid domain state.
var ErrValidation = errors.New("punishment validation failed")

// Violation describes one invalid field.
type Violation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError contains all validation violations.
type ValidationError struct {
	Violations []Violation `json:"violations"`
}

// Error returns the validation error message.
func (err ValidationError) Error() string {
	return ErrValidation.Error()
}

// NewValidationError returns nil when there are no violations.
func NewValidationError(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return ValidationError{Violations: violations}
}

// AddViolation appends a field violation.
func AddViolation(violations []Violation, field string, message string) []Violation {
	return append(violations, Violation{Field: field, Message: message})
}

// ValidateKey validates a stable machine key.
func ValidateKey(field string, key Key) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(string(key))) {
		return []Violation{{Field: field, Message: "must be lower snake case"}}
	}
	return nil
}

// ValidateColor validates a hex color.
func ValidateColor(field string, color Color) []Violation {
	if !colorPattern.MatchString(strings.TrimSpace(string(color))) {
		return []Violation{{Field: field, Message: "must be a hex color"}}
	}
	return nil
}

// ValidateActionKey validates a dotted action key.
func ValidateActionKey(field string, actionKey string) []Violation {
	if !actionPattern.MatchString(strings.TrimSpace(actionKey)) {
		return []Violation{{Field: field, Message: "must be lower dotted words"}}
	}
	return nil
}

// ValidateDefinitionStatus validates definition status.
func ValidateDefinitionStatus(field string, status DefinitionStatus) []Violation {
	if slices.Contains([]DefinitionStatus{DefinitionActive, DefinitionDisabled, DefinitionArchived}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateTargetSystem validates target system.
func ValidateTargetSystem(field string, target TargetSystem) []Violation {
	if slices.Contains([]TargetSystem{TargetGameHub, TargetMinecraft, TargetSAMP, TargetDiscord, TargetCustom}, target) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateEffect validates action effect.
func ValidateEffect(field string, effect Effect) []Violation {
	if slices.Contains([]Effect{EffectRestrict, EffectNotify, EffectExternalEvent}, effect) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateIssuerType validates issuer type.
func ValidateIssuerType(field string, issuerType IssuerType) []Violation {
	if slices.Contains([]IssuerType{IssuerUser, IssuerSystem, IssuerIntegration, IssuerAnticheat, IssuerImport, IssuerUnknown}, issuerType) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidatePunishmentStatus validates punishment status.
func ValidatePunishmentStatus(field string, status PunishmentStatus) []Violation {
	if slices.Contains([]PunishmentStatus{PunishmentActive, PunishmentExpired, PunishmentRevoked, PunishmentVoided, PunishmentPending}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}
