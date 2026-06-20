package application

import (
	"context"
	"testing"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// TestConfigSigningKeysParsesJSONDefaults verifies configured key defaults.
func TestConfigSigningKeysParsesJSONDefaults(t *testing.T) {
	cfg := Config{
		SigningKeysJSON: `[{"key_id":"realmkit:test","public_key":"public-key"}]`,
	}
	keys, err := cfg.SigningKeys()
	if err != nil {
		t.Fatalf("SigningKeys() error = %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("len(keys) = %d, want 1", len(keys))
	}
	key := keys[0]
	if key.Algorithm != domain.SignatureAlgorithmEd25519 {
		t.Fatalf("Algorithm = %q, want ed25519", key.Algorithm)
	}
	if key.TrustLevel != domain.TrustLevelOperator || key.Status != domain.SigningKeyTrusted {
		t.Fatalf("key = %+v, want operator trusted defaults", key)
	}
	if key.Source != domain.SigningKeySourceEnvironment {
		t.Fatalf("Source = %q, want environment", key.Source)
	}
}

// TestSeedSigningKeysUpsertsParsedKeys verifies configured keys reach persistence.
func TestSeedSigningKeysUpsertsParsedKeys(t *testing.T) {
	repository := &fakeSigningKeyRepository{}
	cfg := Config{
		SigningKeysJSON: `[{"key_id":"realmkit:test","public_key":"public-key"}]`,
	}
	if err := SeedSigningKeys(context.Background(), repository, cfg); err != nil {
		t.Fatalf("SeedSigningKeys() error = %v", err)
	}
	if len(repository.keys) != 1 {
		t.Fatalf("len(keys) = %d, want 1", len(repository.keys))
	}
	if repository.keys[0].KeyID != "realmkit:test" {
		t.Fatalf("KeyID = %q, want realmkit:test", repository.keys[0].KeyID)
	}
}

// fakeSigningKeyRepository records signing keys.
type fakeSigningKeyRepository struct {
	keys []domain.ThemeSigningKey
}

// Upsert records a signing key.
func (repository *fakeSigningKeyRepository) Upsert(
	_ context.Context,
	key domain.ThemeSigningKey,
) (domain.ThemeSigningKey, error) {
	repository.keys = append(repository.keys, key)
	return key, nil
}

// FindByKeyID is unused by this fake.
func (repository *fakeSigningKeyRepository) FindByKeyID(context.Context, string) (domain.ThemeSigningKey, error) {
	return domain.ThemeSigningKey{}, nil
}

// List is unused by this fake.
func (repository *fakeSigningKeyRepository) List(context.Context) ([]domain.ThemeSigningKey, error) {
	return repository.keys, nil
}
