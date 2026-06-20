package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// testActivation returns a public activation.
func testActivation(themeID uuid.UUID, versionID uuid.UUID) domain.ThemeActivation {
	return domain.ThemeActivation{
		ID:               uuid.New(),
		ThemeID:          themeID,
		VersionID:        versionID,
		Environment:      domain.EnvironmentPublic,
		Reason:           "Initial publication",
		SettingsDataJSON: []byte(`{"accent":"lime"}`),
	}
}

// testSigningKey returns a trusted signing key.
func testSigningKey() domain.ThemeSigningKey {
	return domain.ThemeSigningKey{
		ID:          uuid.New(),
		KeyID:       "realmkit:test",
		Algorithm:   domain.SignatureAlgorithmEd25519,
		PublicKey:   "public-key",
		TrustLevel:  domain.TrustLevelOperator,
		Status:      domain.SigningKeyTrusted,
		Source:      domain.SigningKeySourceEnvironment,
		Description: "Test key",
	}
}

// testPreviewToken returns a preview token.
func testPreviewToken(versionID uuid.UUID) domain.ThemePreviewToken {
	return domain.ThemePreviewToken{
		ID:            uuid.New(),
		VersionID:     versionID,
		TokenHash:     "preview-token-hash",
		PersonaKind:   domain.PersonaModerator,
		PersonaSource: domain.PersonaSourceSynthetic,
		ExpiresAt:     time.Now().UTC().Add(time.Hour),
	}
}

// integrityFiles converts stored files into integrity hash inputs.
func integrityFiles(files []domain.ThemeFile) []domain.IntegrityFile {
	inputs := make([]domain.IntegrityFile, 0, len(files))
	for _, file := range files {
		inputs = append(inputs, domain.IntegrityFile{Path: file.Path, ContentSHA256: file.ContentSHA256})
	}
	return inputs
}
