package signing

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Clock returns the current time.
type Clock func() time.Time

// Verifier verifies detached package signatures.
type Verifier struct {
	keys  port.SigningKeyRepository // keys stores the keys value.
	cfg   Config                    // cfg stores the cfg value.
	clock Clock                     // clock stores the clock value.
}

// Result is the signature verification result.
type Result struct {
	Signature domain.ThemePackageSignature  // Signature stores the signature value.
	Issues    []domain.ThemeValidationIssue // Issues stores the issues value.
}

// NewVerifier creates a package signature verifier.
func NewVerifier(keys port.SigningKeyRepository, cfg Config, clock Clock) Verifier {
	if clock == nil {
		clock = time.Now
	}
	return Verifier{keys: keys, cfg: cfg, clock: clock}
}

// Verify verifies manifest and signature bytes.
func (verifier Verifier) Verify(
	ctx context.Context,
	manifest []byte,
	signature []byte,
) Result {
	if len(signature) == 0 {
		return verifier.missingSignature()
	}
	envelope, err := DecodeEnvelope(signature)
	if err != nil {
		return invalidResult("", "", "Malformed package signature envelope.")
	}
	signedAt, err := parseSignedAt(envelope.SignedAt)
	if err != nil {
		return invalidResult(envelope.KeyID, envelope.Signature, "Package signature signed_at is invalid.")
	}
	return verifier.verifyEnvelope(ctx, manifest, envelope, signedAt)
}

// missingSignature returns the configured unsigned-package result.
func (verifier Verifier) missingSignature() Result {
	severity := domain.SeverityError
	if verifier.cfg.AllowUnsignedPackages {
		severity = domain.SeverityWarning
	}
	return Result{
		Signature: domain.ThemePackageSignature{
			ID:                 uuid.New(),
			Algorithm:          domain.SignatureAlgorithmEd25519,
			VerificationStatus: domain.SignatureMissing,
		},
		Issues: []domain.ThemeValidationIssue{issue(severity, domain.IssueMissingSignature, "Package signature is missing.")},
	}
}

// verifyEnvelope verifies one decoded signature envelope.
func (verifier Verifier) verifyEnvelope(
	ctx context.Context,
	manifest []byte,
	envelope Envelope,
	signedAt time.Time,
) Result {
	digest, canonical, err := manifestDigestAndCanonical(manifest)
	if err != nil {
		return invalidResult(envelope.KeyID, envelope.Signature, "Theme manifest is malformed.")
	}
	if NormalizeDigest(envelope.ManifestSHA256) != digest {
		return invalidResult(envelope.KeyID, envelope.Signature, "Package signature manifest hash does not match.")
	}
	key, err := verifier.keys.FindByKeyID(ctx, envelope.KeyID)
	if err != nil {
		return keyIssue(envelope, digest, domain.SignatureUntrusted, domain.IssueUntrustedSignature, "Package signing key is not trusted.")
	}
	return verifier.verifyWithKey(key, envelope, digest, canonical, signedAt)
}

// verifyWithKey verifies bytes with one trusted key and policy state.
func (verifier Verifier) verifyWithKey(
	key domain.ThemeSigningKey,
	envelope Envelope,
	digest domain.Digest,
	canonical []byte,
	signedAt time.Time,
) Result {
	if key.Status == domain.SigningKeyRevoked || key.RevokedAt != nil {
		return keyIssue(envelope, digest, domain.SignatureRevoked, domain.IssueRevokedSignature, "Package signing key is revoked.")
	}
	publicKey, decoded := decodePublicKey(key.PublicKey)
	signature, err := decodeBase64(envelope.Signature)
	if !decoded || err != nil || !ed25519.Verify(publicKey, canonical, signature) {
		return invalidResult(envelope.KeyID, envelope.Signature, "Package signature verification failed.")
	}
	if !withinKeyWindow(key, signedAt) {
		return invalidResult(envelope.KeyID, envelope.Signature, "Package signature timestamp is outside the key validity window.")
	}
	if key.Status == domain.SigningKeyRetired && retiredAfterCutoff(key, signedAt) {
		return keyIssue(envelope, digest, domain.SignatureRetired, domain.IssueRetiredSignature, "Package was signed after the key retired.")
	}
	status := domain.SignatureVerified
	if key.Status == domain.SigningKeyRetired {
		status = domain.SignatureRetired
	}
	verifiedAt := verifier.clock().UTC()
	return Result{Signature: signatureRecord(envelope, digest, status, &verifiedAt)}
}

// manifestDigestAndCanonical returns the canonical manifest and digest.
func manifestDigestAndCanonical(manifest []byte) (domain.Digest, []byte, error) {
	canonical, err := CanonicalManifestJSON(manifest)
	if err != nil {
		return "", nil, err
	}
	hash := sha256Bytes(canonical)
	return hash, canonical, nil
}

// sha256Bytes returns a lowercase SHA-256 digest.
func sha256Bytes(value []byte) domain.Digest {
	hash := sha256.Sum256(value)
	return domain.Digest(hex.EncodeToString(hash[:]))
}
