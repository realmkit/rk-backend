package signing

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// TestVerifierAcceptsTrustedSignature verifies valid Ed25519 envelopes.
func TestVerifierAcceptsTrustedSignature(t *testing.T) {
	manifest := []byte(`{"version":"1.0.0","name":"Main"}`)
	key, signature := signedPackage(t, manifest, domain.SigningKeyTrusted, nil)
	verifier := NewVerifier(fakeSigningKeyRepository{key: key}, Config{}, fixedClock())
	result := verifier.Verify(context.Background(), manifest, signature)
	if result.Signature.VerificationStatus != domain.SignatureVerified {
		t.Fatalf("VerificationStatus = %q, want verified", result.Signature.VerificationStatus)
	}
	if len(result.Issues) != 0 {
		t.Fatalf("Issues = %+v, want none", result.Issues)
	}
	if result.Signature.VerifiedAt == nil {
		t.Fatalf("VerifiedAt = nil, want timestamp")
	}
}

// TestVerifierReportsMissingSignaturePolicy verifies unsigned package policy.
func TestVerifierReportsMissingSignaturePolicy(t *testing.T) {
	strict := NewVerifier(fakeSigningKeyRepository{}, Config{}, fixedClock())
	strictResult := strict.Verify(context.Background(), []byte(`{"name":"Main"}`), nil)
	if strictResult.Issues[0].Severity != domain.SeverityError {
		t.Fatalf("strict severity = %q, want error", strictResult.Issues[0].Severity)
	}
	local := NewVerifier(fakeSigningKeyRepository{}, Config{AllowUnsignedPackages: true}, fixedClock())
	localResult := local.Verify(context.Background(), []byte(`{"name":"Main"}`), nil)
	if localResult.Issues[0].Severity != domain.SeverityWarning {
		t.Fatalf("local severity = %q, want warning", localResult.Issues[0].Severity)
	}
}

// TestVerifierRejectsTamperedManifest verifies manifest hash protection.
func TestVerifierRejectsTamperedManifest(t *testing.T) {
	manifest := []byte(`{"version":"1.0.0","name":"Main"}`)
	key, signature := signedPackage(t, manifest, domain.SigningKeyTrusted, nil)
	verifier := NewVerifier(fakeSigningKeyRepository{key: key}, Config{}, fixedClock())
	result := verifier.Verify(context.Background(), []byte(`{"version":"1.0.1","name":"Main"}`), signature)
	if result.Signature.VerificationStatus != domain.SignatureInvalid {
		t.Fatalf("VerificationStatus = %q, want invalid", result.Signature.VerificationStatus)
	}
	if result.Issues[0].Code != domain.IssueInvalidSignature {
		t.Fatalf("Issue code = %q, want invalid signature", result.Issues[0].Code)
	}
}

// TestVerifierAppliesKeyTrustPolicy verifies untrusted, retired, and revoked keys.
func TestVerifierAppliesKeyTrustPolicy(t *testing.T) {
	manifest := []byte(`{"name":"Main"}`)
	retiredAt := fixedNow().Add(time.Hour)
	retiredKey, retiredSignature := signedPackage(t, manifest, domain.SigningKeyRetired, &retiredAt)
	retired := NewVerifier(fakeSigningKeyRepository{key: retiredKey}, Config{}, fixedClock())
	retiredResult := retired.Verify(context.Background(), manifest, retiredSignature)
	if retiredResult.Signature.VerificationStatus != domain.SignatureRetired || len(retiredResult.Issues) != 0 {
		t.Fatalf("retired result = %+v, want accepted retired", retiredResult)
	}
	revokedKey, revokedSignature := signedPackage(t, manifest, domain.SigningKeyRevoked, nil)
	revoked := NewVerifier(fakeSigningKeyRepository{key: revokedKey}, Config{}, fixedClock())
	revokedResult := revoked.Verify(context.Background(), manifest, revokedSignature)
	if revokedResult.Issues[0].Code != domain.IssueRevokedSignature {
		t.Fatalf("revoked issue = %q, want revoked", revokedResult.Issues[0].Code)
	}
	untrusted := NewVerifier(fakeSigningKeyRepository{}, Config{}, fixedClock())
	untrustedResult := untrusted.Verify(context.Background(), manifest, retiredSignature)
	if untrustedResult.Issues[0].Code != domain.IssueUntrustedSignature {
		t.Fatalf("untrusted issue = %q, want untrusted", untrustedResult.Issues[0].Code)
	}
}

// fakeSigningKeyRepository returns one configured signing key.
type fakeSigningKeyRepository struct {
	key domain.ThemeSigningKey
}

// Upsert is unused by verifier tests.
func (repository fakeSigningKeyRepository) Upsert(
	context.Context,
	domain.ThemeSigningKey,
) (domain.ThemeSigningKey, error) {
	return domain.ThemeSigningKey{}, nil
}

// FindByKeyID returns the configured signing key.
func (repository fakeSigningKeyRepository) FindByKeyID(
	_ context.Context,
	keyID string,
) (domain.ThemeSigningKey, error) {
	if repository.key.KeyID != keyID {
		return domain.ThemeSigningKey{}, portErrNotFound()
	}
	return repository.key, nil
}

// List is unused by verifier tests.
func (repository fakeSigningKeyRepository) List(context.Context) ([]domain.ThemeSigningKey, error) {
	return nil, nil
}

// signedPackage returns a key and envelope for manifest bytes.
func signedPackage(
	t *testing.T,
	manifest []byte,
	status domain.SigningKeyStatus,
	retiredAt *time.Time,
) (domain.ThemeSigningKey, []byte) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	canonical, err := CanonicalManifestJSON(manifest)
	if err != nil {
		t.Fatalf("CanonicalManifestJSON() error = %v", err)
	}
	digest, err := ManifestSHA256(manifest)
	if err != nil {
		t.Fatalf("ManifestSHA256() error = %v", err)
	}
	signature := ed25519.Sign(privateKey, canonical)
	envelope, err := json.Marshal(Envelope{
		Algorithm:      string(domain.SignatureAlgorithmEd25519),
		KeyID:          "realmkit:test",
		ManifestSHA256: string(digest),
		Signature:      base64.RawURLEncoding.EncodeToString(signature),
		SignedAt:       fixedNow().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Marshal(envelope) error = %v", err)
	}
	return domain.ThemeSigningKey{
		KeyID:      "realmkit:test",
		Algorithm:  domain.SignatureAlgorithmEd25519,
		PublicKey:  base64.RawURLEncoding.EncodeToString(publicKey),
		TrustLevel: domain.TrustLevelOperator,
		Status:     status,
		Source:     domain.SigningKeySourceDatabase,
		RetiredAt:  retiredAt,
	}, envelope
}

// fixedClock returns a deterministic verifier clock.
func fixedClock() Clock {
	return func() time.Time { return fixedNow() }
}

// fixedNow returns the test timestamp.
func fixedNow() time.Time {
	return time.Date(2026, time.June, 19, 15, 0, 0, 0, time.UTC)
}

// portErrNotFound returns the port not found error without exporting test helpers.
func portErrNotFound() error {
	return port.ErrNotFound
}
