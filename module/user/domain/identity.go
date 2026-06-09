package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// IdentityLink links a local user to an identity provider subject.
type IdentityLink struct {
	// ID is the link identifier.
	ID uuid.UUID `json:"id"`

	// UserID is the local user identifier.
	UserID uuid.UUID `json:"user_id"`

	// Provider is the configured provider preset.
	Provider string `json:"provider"`

	// Issuer is the OIDC issuer.
	Issuer string `json:"issuer"`

	// Subject is the provider subject.
	Subject string `json:"subject"`

	// SubjectHash is the log-safe subject hash.
	SubjectHash string `json:"subject_hash"`

	// ClaimsHash is the latest claim cache hash.
	ClaimsHash string `json:"claims_hash"`

	// LinkedAt is when the identity was linked.
	LinkedAt time.Time `json:"linked_at"`

	// LastSeenAt is when the identity last authenticated.
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`

	// LastSyncedAt is when provider claims were last synced.
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

// ClaimCache stores read-only provider-owned display claims.
type ClaimCache struct {
	// ID is the cache row identifier.
	ID uuid.UUID `json:"id"`

	// UserID is the local user identifier.
	UserID uuid.UUID `json:"user_id"`

	// Issuer is the OIDC issuer.
	Issuer string `json:"issuer"`

	// Subject is the provider subject.
	Subject string `json:"subject"`

	// Username is provider-owned display data.
	Username string `json:"username"`

	// Email is provider-owned contact data.
	Email string `json:"email"`

	// EmailVerified reports whether the provider verified the email.
	EmailVerified bool `json:"email_verified"`

	// DisplayName is provider-owned display data.
	DisplayName string `json:"display_name"`

	// PictureURL is the provider picture fallback.
	PictureURL string `json:"picture_url"`

	// PreferredLocale is the provider locale.
	PreferredLocale string `json:"preferred_locale"`

	// ClaimsHash is the normalized claims hash.
	ClaimsHash string `json:"claims_hash"`

	// SyncedAt is the sync timestamp.
	SyncedAt time.Time `json:"synced_at"`
}

// Validate validates identity link fields.
func (link IdentityLink) Validate() error {
	var violations []Violation
	if link.UserID == uuid.Nil {
		violations = AppendViolation(violations, "user_id", "is required")
	}
	if strings.TrimSpace(link.Provider) == "" {
		violations = AppendViolation(violations, "provider", "is required")
	}
	if strings.TrimSpace(link.Issuer) == "" {
		violations = AppendViolation(violations, "issuer", "is required")
	}
	if strings.TrimSpace(link.Subject) == "" {
		violations = AppendViolation(violations, "subject", "is required")
	}
	if strings.TrimSpace(link.SubjectHash) == "" {
		violations = AppendViolation(violations, "subject_hash", "is required")
	}
	return NewValidationError(violations)
}
