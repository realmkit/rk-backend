package signing

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// issue creates a structured package validation issue.
func issue(
	severity domain.ValidationSeverity,
	code domain.ValidationIssueCode,
	message string,
) domain.ThemeValidationIssue {
	return domain.ThemeValidationIssue{
		ID:       uuid.New(),
		Severity: severity,
		Code:     code,
		Message:  message,
		Details:  []byte(`{}`),
	}
}

// invalidResult returns an invalid signature result.
func invalidResult(keyID string, signature string, message string) Result {
	return Result{
		Signature: signatureRecord(
			Envelope{KeyID: keyID, Signature: signature},
			"",
			domain.SignatureInvalid,
			nil,
		),
		Issues: []domain.ThemeValidationIssue{issue(domain.SeverityError, domain.IssueInvalidSignature, message)},
	}
}

// keyIssue returns a trust-policy signature result.
func keyIssue(
	envelope Envelope,
	digest domain.Digest,
	status domain.SignatureVerificationStatus,
	code domain.ValidationIssueCode,
	message string,
) Result {
	return Result{
		Signature: signatureRecord(envelope, digest, status, nil),
		Issues:    []domain.ThemeValidationIssue{issue(domain.SeverityError, code, message)},
	}
}

// signatureRecord maps a verification envelope into domain persistence.
func signatureRecord(
	envelope Envelope,
	digest domain.Digest,
	status domain.SignatureVerificationStatus,
	verifiedAt *time.Time,
) domain.ThemePackageSignature {
	return domain.ThemePackageSignature{
		ID:                 uuid.New(),
		KeyID:              envelope.KeyID,
		Algorithm:          domain.SignatureAlgorithmEd25519,
		VerificationStatus: status,
		Signature:          envelope.Signature,
		SignedManifestHash: digest,
		VerifiedAt:         verifiedAt,
	}
}

// parseSignedAt parses an RFC3339 signature timestamp.
func parseSignedAt(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}

// decodePublicKey decodes an Ed25519 public key.
func decodePublicKey(value string) (ed25519.PublicKey, bool) {
	decoded, err := decodeBase64(value)
	if err != nil {
		decoded, err = hex.DecodeString(value)
	}
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, false
	}
	return ed25519.PublicKey(decoded), true
}

// decodeBase64 decodes standard or URL-safe base64 text.
func decodeBase64(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	decoded, err = base64.StdEncoding.DecodeString(value)
	if err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(value)
}

// withinKeyWindow reports whether a signature timestamp is key-valid.
func withinKeyWindow(key domain.ThemeSigningKey, signedAt time.Time) bool {
	if signedAt.IsZero() {
		return true
	}
	if key.NotBefore != nil && signedAt.Before(*key.NotBefore) {
		return false
	}
	if key.NotAfter != nil && signedAt.After(*key.NotAfter) {
		return false
	}
	return true
}

// retiredAfterCutoff reports whether a retired key signed too late.
func retiredAfterCutoff(key domain.ThemeSigningKey, signedAt time.Time) bool {
	if key.RetiredAt == nil || signedAt.IsZero() {
		return false
	}
	return signedAt.After(*key.RetiredAt)
}
