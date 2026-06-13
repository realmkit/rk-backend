package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// ExternalIdentity contains provider-neutral user claim data.
type ExternalIdentity struct {
	// Issuer is the OIDC token issuer.
	Issuer string

	// Subject is the stable provider subject.
	Subject string

	// Username is provider-owned display data.
	Username string

	// Email is provider-owned contact data.
	Email string

	// EmailVerified reports whether the provider verified the email.
	EmailVerified bool

	// DisplayName is provider-owned display data.
	DisplayName string

	// PictureURL is the provider picture fallback.
	PictureURL string

	// PreferredLocale is the provider locale claim.
	PreferredLocale string

	// RawClaimsHash is a hash of normalized raw claims.
	RawClaimsHash string
}

// FromClaims maps JWT claims into an external identity.
func FromClaims(claims map[string]any) (ExternalIdentity, error) {
	issuer := stringClaim(claims, "iss")
	subject := stringClaim(claims, "sub")
	if issuer == "" {
		return ExternalIdentity{}, fmt.Errorf("issuer claim is required")
	}
	if subject == "" {
		return ExternalIdentity{}, fmt.Errorf("subject claim is required")
	}
	return ExternalIdentity{
		Issuer:          issuer,
		Subject:         subject,
		Username:        firstStringClaim(claims, "preferred_username", "username", "nickname"),
		Email:           stringClaim(claims, "email"),
		EmailVerified:   boolClaim(claims, "email_verified"),
		DisplayName:     firstStringClaim(claims, "name", "display_name"),
		PictureURL:      stringClaim(claims, "picture"),
		PreferredLocale: stringClaim(claims, "locale"),
		RawClaimsHash:   ClaimsHash(claims),
	}, nil
}

// Merge fills optional provider profile claims from another identity.
func (external ExternalIdentity) Merge(other ExternalIdentity) ExternalIdentity {
	if external.Username == "" {
		external.Username = other.Username
	}
	if external.Email == "" {
		external.Email = other.Email
		external.EmailVerified = other.EmailVerified
	}
	if external.DisplayName == "" {
		external.DisplayName = other.DisplayName
	}
	if external.PictureURL == "" {
		external.PictureURL = other.PictureURL
	}
	if external.PreferredLocale == "" {
		external.PreferredLocale = other.PreferredLocale
	}
	if other.RawClaimsHash != "" {
		external.RawClaimsHash = other.RawClaimsHash
	}
	return external
}

// SubjectHash returns a log-safe identity key hash.
func SubjectHash(issuer string, subject string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(issuer) + "\x00" + strings.TrimSpace(subject)))
	return hex.EncodeToString(sum[:])
}

// ClaimsHash returns a stable hash for claims.
func ClaimsHash(claims map[string]any) string {
	body, _ := json.Marshal(claims)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// stringClaim returns a trimmed string claim.
func stringClaim(claims map[string]any, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}

// firstStringClaim returns the first nonblank string claim.
func firstStringClaim(claims map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringClaim(claims, key); value != "" {
			return value
		}
	}
	return ""
}

// boolClaim returns a boolean claim.
func boolClaim(claims map[string]any, key string) bool {
	value, _ := claims[key].(bool)
	return value
}
