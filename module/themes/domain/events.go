package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// EventKey identifies one theme event fact.
type EventKey string

const (
	// EventThemeCreated is emitted when a theme family is created.
	EventThemeCreated EventKey = "themes.theme.created"
	// EventThemeUpdated is emitted when theme metadata changes.
	EventThemeUpdated EventKey = "themes.theme.updated"
	// EventVersionImported is emitted when a package import is persisted.
	EventVersionImported EventKey = "themes.version.imported"
	// EventVersionValidated is emitted when validation completes.
	EventVersionValidated EventKey = "themes.version.validated"
	// EventVersionFileSaved is emitted when a draft file changes.
	EventVersionFileSaved EventKey = "themes.version.file_saved"
	// EventVersionArchived is emitted when a version is archived.
	EventVersionArchived EventKey = "themes.version.archived"
	// EventActivationChanged is emitted when public or preview activation changes.
	EventActivationChanged EventKey = "themes.activation.changed"
	// EventActivationRolledBack is emitted when rollback creates a new activation.
	EventActivationRolledBack EventKey = "themes.activation.rolled_back"
	// EventSigningKeyCreated is emitted when an operator adds a signing key.
	EventSigningKeyCreated EventKey = "themes.signing_key.created"
	// EventSigningKeyRetired is emitted when a key is retired.
	EventSigningKeyRetired EventKey = "themes.signing_key.retired"
	// EventSigningKeyRevoked is emitted when a key is revoked.
	EventSigningKeyRevoked EventKey = "themes.signing_key.revoked"
	// EventCacheInvalidated is emitted when theme delivery caches must refresh.
	EventCacheInvalidated EventKey = "themes.cache.invalidated"
)

// ThemeEventKeys returns all built-in theme event keys.
func ThemeEventKeys() []EventKey {
	return []EventKey{
		EventThemeCreated,
		EventThemeUpdated,
		EventVersionImported,
		EventVersionValidated,
		EventVersionFileSaved,
		EventVersionArchived,
		EventActivationChanged,
		EventActivationRolledBack,
		EventSigningKeyCreated,
		EventSigningKeyRetired,
		EventSigningKeyRevoked,
		EventCacheInvalidated,
	}
}

// EnsurePublishable returns an error when a version cannot be activated.
func (version ThemeVersion) EnsurePublishable(
	signature ThemePackageSignature,
	issues []ThemeValidationIssue,
) error {
	if version.Status != VersionStatusValid && version.Status != VersionStatusPublished {
		return ErrVersionNotPublishable
	}
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return ErrVersionNotPublishable
		}
	}
	if signature.VerificationStatus != SignatureVerified && signature.VerificationStatus != SignatureRetired {
		return ErrVersionNotPublishable
	}
	return nil
}

// MarkPublished returns a copy marked as publicly published.
func (version ThemeVersion) MarkPublished(publishedAt time.Time, actorUserID *uuid.UUID) (ThemeVersion, error) {
	if version.Status != VersionStatusValid && version.Status != VersionStatusPublished {
		return ThemeVersion{}, ErrVersionNotPublishable
	}
	next := version
	at := publishedAt.UTC()
	next.Status = VersionStatusPublished
	next.PublishedAt = &at
	next.PublishedBy = actorUserID
	next.UpdatedBy = actorUserID
	return next, nil
}

// Validate returns an error when an activation cannot be persisted.
func (activation ThemeActivation) Validate() error {
	if activation.ThemeID == uuid.Nil || activation.VersionID == uuid.Nil {
		return ErrInvalidActivation
	}
	if activation.Environment != EnvironmentPublic && activation.Environment != EnvironmentPreview {
		return ErrInvalidActivation
	}
	if activation.ActivatedAt.IsZero() {
		return ErrInvalidActivation
	}
	return nil
}

// EnsureUsableAt returns an error when a signing key is inactive at a time.
func (key ThemeSigningKey) EnsureUsableAt(now time.Time) error {
	if key.Status == SigningKeyRevoked || key.RevokedAt != nil {
		return ErrSigningKeyInactive
	}
	if key.Status != SigningKeyTrusted && key.Status != SigningKeyRetired {
		return ErrSigningKeyInactive
	}
	if key.NotBefore != nil && now.Before(*key.NotBefore) {
		return ErrSigningKeyInactive
	}
	if key.NotAfter != nil && now.After(*key.NotAfter) {
		return ErrSigningKeyInactive
	}
	return nil
}

// Retire returns a copy of a signing key marked retired.
func (key ThemeSigningKey) Retire(retiredAt time.Time) (ThemeSigningKey, error) {
	if key.Status == SigningKeyRevoked || key.RevokedAt != nil {
		return ThemeSigningKey{}, ErrSigningKeyInactive
	}
	next := key
	at := retiredAt.UTC()
	next.Status = SigningKeyRetired
	next.RetiredAt = &at
	next.UpdatedAt = at
	return next, nil
}

// Revoke returns a copy of a signing key marked revoked.
func (key ThemeSigningKey) Revoke(revokedAt time.Time) ThemeSigningKey {
	next := key
	at := revokedAt.UTC()
	next.Status = SigningKeyRevoked
	next.RevokedAt = &at
	next.UpdatedAt = at
	return next
}

// ValidateAt returns an error when a preview token is unusable.
func (token ThemePreviewToken) ValidateAt(now time.Time) error {
	if token.VersionID == uuid.Nil || token.TokenHash == "" {
		return ErrPreviewTokenInvalid
	}
	if token.RevokedAt != nil {
		return ErrPreviewTokenRevoked
	}
	if !token.ExpiresAt.After(now) {
		return ErrPreviewTokenExpired
	}
	return nil
}

// Revoke returns a copy of a preview token marked revoked.
func (token ThemePreviewToken) Revoke(revokedAt time.Time) ThemePreviewToken {
	if token.RevokedAt != nil {
		return token
	}
	next := token
	at := revokedAt.UTC()
	next.RevokedAt = &at
	return next
}

// IsPreviewTokenStateError reports whether an error came from preview token state.
func IsPreviewTokenStateError(err error) bool {
	return errors.Is(err, ErrPreviewTokenInvalid) ||
		errors.Is(err, ErrPreviewTokenExpired) ||
		errors.Is(err, ErrPreviewTokenRevoked)
}
