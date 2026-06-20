package delivery

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// fakeActivationRepository stores activations in memory.
type fakeActivationRepository struct {
	current domain.ThemeActivation
}

// Activate stores one activation.
func (repository fakeActivationRepository) Activate(context.Context, domain.ThemeActivation) (domain.ThemeActivation, error) {
	return domain.ThemeActivation{}, nil
}

// Current returns the current activation.
func (repository fakeActivationRepository) Current(context.Context, domain.ActivationEnvironment) (domain.ThemeActivation, error) {
	return repository.current, nil
}

// FindByID returns no activation.
func (repository fakeActivationRepository) FindByID(context.Context, uuid.UUID) (domain.ThemeActivation, error) {
	return domain.ThemeActivation{}, port.ErrNotFound
}

// ListByTheme returns no activations.
func (repository fakeActivationRepository) ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeActivation, error) {
	return nil, nil
}

// fakeIssueRepository stores issues in memory.
type fakeIssueRepository struct {
	issues map[uuid.UUID][]domain.ThemeValidationIssue
}

// ReplaceVersionIssues replaces issues.
func (repository fakeIssueRepository) ReplaceVersionIssues(context.Context, uuid.UUID, []domain.ThemeValidationIssue) error {
	return nil
}

// ListByVersion returns issues.
func (repository fakeIssueRepository) ListByVersion(_ context.Context, versionID uuid.UUID) ([]domain.ThemeValidationIssue, error) {
	return repository.issues[versionID], nil
}

// fakePreviewTokenRepository stores preview tokens in memory.
type fakePreviewTokenRepository struct {
	tokens map[string]domain.ThemePreviewToken
}

// Create stores one preview token.
func (repository fakePreviewTokenRepository) Create(
	_ context.Context,
	token domain.ThemePreviewToken,
) (domain.ThemePreviewToken, error) {
	repository.tokens[token.TokenHash] = token
	return token, nil
}

// FindByTokenHash returns one token.
func (repository fakePreviewTokenRepository) FindByTokenHash(
	_ context.Context,
	tokenHash string,
) (domain.ThemePreviewToken, error) {
	token, ok := repository.tokens[tokenHash]
	if !ok || token.RevokedAt != nil {
		return domain.ThemePreviewToken{}, port.ErrNotFound
	}
	return token, nil
}

// Revoke revokes one token.
func (repository fakePreviewTokenRepository) Revoke(_ context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	for key, token := range repository.tokens {
		if token.ID == id {
			token.RevokedAt = &now
			repository.tokens[key] = token
			return nil
		}
	}
	return port.ErrNotFound
}
