package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Config contains theme runtime settings.
type Config struct {
	// SigningKeysJSON is a JSON array of trusted package signing keys.
	SigningKeysJSON string `mapstructure:"signing_keys_json" default:"[]"`
}

// SigningKeySeed is one configured package signing key.
type SigningKeySeed struct {
	KeyID       string `json:"key_id"`
	Algorithm   string `json:"algorithm"`
	PublicKey   string `json:"public_key"`
	TrustLevel  string `json:"trust_level"`
	Status      string `json:"status"`
	NotBefore   string `json:"not_before"`
	NotAfter    string `json:"not_after"`
	Description string `json:"description"`
}

// SeedSigningKeys upserts configured signing keys into persistence.
func SeedSigningKeys(ctx context.Context, repository port.SigningKeyRepository, cfg Config) error {
	keys, err := cfg.SigningKeys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if _, err := repository.Upsert(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// SigningKeys parses configured signing keys.
func (cfg Config) SigningKeys() ([]domain.ThemeSigningKey, error) {
	source := strings.TrimSpace(cfg.SigningKeysJSON)
	if source == "" {
		source = "[]"
	}
	var seeds []SigningKeySeed
	if err := json.Unmarshal([]byte(source), &seeds); err != nil {
		return nil, fmt.Errorf("parse theme signing keys: %w", err)
	}
	keys := make([]domain.ThemeSigningKey, 0, len(seeds))
	for _, seed := range seeds {
		key, err := seed.SigningKey()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// SigningKey converts a configured seed into a domain signing key.
func (seed SigningKeySeed) SigningKey() (domain.ThemeSigningKey, error) {
	if strings.TrimSpace(seed.KeyID) == "" {
		return domain.ThemeSigningKey{}, fmt.Errorf("theme signing key id is required")
	}
	if strings.TrimSpace(seed.PublicKey) == "" {
		return domain.ThemeSigningKey{}, fmt.Errorf("theme signing key public key is required")
	}
	notBefore, err := parseOptionalTime(seed.NotBefore)
	if err != nil {
		return domain.ThemeSigningKey{}, fmt.Errorf("parse theme signing key not_before: %w", err)
	}
	notAfter, err := parseOptionalTime(seed.NotAfter)
	if err != nil {
		return domain.ThemeSigningKey{}, fmt.Errorf("parse theme signing key not_after: %w", err)
	}
	return domain.ThemeSigningKey{
		KeyID:       strings.TrimSpace(seed.KeyID),
		Algorithm:   defaultAlgorithm(seed.Algorithm),
		PublicKey:   strings.TrimSpace(seed.PublicKey),
		TrustLevel:  defaultTrustLevel(seed.TrustLevel),
		Status:      defaultStatus(seed.Status),
		Source:      domain.SigningKeySourceEnvironment,
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		Description: seed.Description,
	}, nil
}

// parseOptionalTime parses optional RFC3339 timestamps.
func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// defaultAlgorithm returns the configured or default signature algorithm.
func defaultAlgorithm(value string) domain.SignatureAlgorithm {
	if value == "" {
		return domain.SignatureAlgorithmEd25519
	}
	return domain.SignatureAlgorithm(value)
}

// defaultTrustLevel returns the configured or default trust level.
func defaultTrustLevel(value string) domain.SigningKeyTrustLevel {
	if value == "" {
		return domain.TrustLevelOperator
	}
	return domain.SigningKeyTrustLevel(value)
}

// defaultStatus returns the configured or default signing key status.
func defaultStatus(value string) domain.SigningKeyStatus {
	if value == "" {
		return domain.SigningKeyTrusted
	}
	return domain.SigningKeyStatus(value)
}
