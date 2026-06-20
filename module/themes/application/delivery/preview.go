package delivery

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

const (
	// defaultPreviewTTL is used when callers do not request a duration.
	defaultPreviewTTL = 15 * time.Minute
	// maxPreviewTTL is the longest first-version preview token duration.
	maxPreviewTTL = 24 * time.Hour
)

// CreatePreviewToken creates a scoped expiring preview token.
func (service Service) CreatePreviewToken(
	ctx context.Context,
	command CreatePreviewTokenCommand,
) (PreviewTokenResult, error) {
	version, err := service.repositories.Versions.FindByID(ctx, command.VersionID)
	if err != nil {
		return PreviewTokenResult{}, err
	}
	rawToken, err := randomToken()
	if err != nil {
		return PreviewTokenResult{}, err
	}
	expiresAt := service.clock().UTC().Add(previewTTL(command.TTL))
	token := domain.ThemePreviewToken{
		ID:            uuid.New(),
		VersionID:     version.ID,
		TokenHash:     tokenHash(rawToken),
		PersonaKind:   previewPersona(command.PersonaKind),
		PersonaSource: previewSource(command.PersonaSource),
		PersonaUserID: command.PersonaUserID,
		ExpiresAt:     expiresAt,
		CreatedBy:     command.ActorUserID,
	}
	stored, err := service.repositories.PreviewTokens.Create(ctx, token)
	if err != nil {
		return PreviewTokenResult{}, err
	}
	return PreviewTokenResult{Token: rawToken, Preview: stored, ExpiresAt: expiresAt}, nil
}

// ValidatePreviewToken validates and returns an active preview token.
func (service Service) ValidatePreviewToken(
	ctx context.Context,
	command ValidatePreviewTokenCommand,
) (domain.ThemePreviewToken, error) {
	token, err := service.repositories.PreviewTokens.FindByTokenHash(ctx, tokenHash(command.Token))
	if err != nil {
		return domain.ThemePreviewToken{}, err
	}
	if !token.ExpiresAt.After(service.clock().UTC()) {
		return domain.ThemePreviewToken{}, port.ErrInvalidState
	}
	return token, nil
}

// randomToken creates a URL-safe preview bearer token.
func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// tokenHash returns a stable token hash for storage lookup.
func tokenHash(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// previewTTL clamps preview token duration.
func previewTTL(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultPreviewTTL
	}
	if value > maxPreviewTTL {
		return maxPreviewTTL
	}
	return value
}

// previewPersona returns a safe default persona.
func previewPersona(value domain.PreviewPersonaKind) domain.PreviewPersonaKind {
	if value == "" {
		return domain.PersonaAnonymous
	}
	return value
}

// previewSource returns a safe default persona source.
func previewSource(value domain.PreviewPersonaSource) domain.PreviewPersonaSource {
	if value == "" {
		return domain.PersonaSourceSynthetic
	}
	return value
}
