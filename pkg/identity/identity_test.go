package identity

import "testing"

// TestFromClaimsMapsProviderData verifies provider claims normalize into identity data.
func TestFromClaimsMapsProviderData(t *testing.T) {
	claims := map[string]any{
		"iss":                "https://auth.example",
		"sub":                "subject",
		"preferred_username": "ian",
		"email":              "ian@example.com",
		"email_verified":     true,
		"name":               "Ian",
		"picture":            "https://cdn.example/avatar.png",
		"locale":             "en",
	}

	identity, err := FromClaims(claims)
	if err != nil {
		t.Fatalf("FromClaims() error = %v", err)
	}
	if identity.Issuer != "https://auth.example" || identity.Subject != "subject" || identity.Username != "ian" || !identity.EmailVerified {
		t.Fatalf("identity = %+v, want mapped claims", identity)
	}
	if identity.RawClaimsHash == "" || SubjectHash(identity.Issuer, identity.Subject) == "" {
		t.Fatalf("hashes must not be empty")
	}
}

// TestFromClaimsRequiresIssuerAndSubject verifies identity keys are mandatory.
func TestFromClaimsRequiresIssuerAndSubject(t *testing.T) {
	if _, err := FromClaims(map[string]any{"iss": "issuer"}); err == nil {
		t.Fatalf("FromClaims() error = nil, want missing subject error")
	}
	if _, err := FromClaims(map[string]any{"sub": "subject"}); err == nil {
		t.Fatalf("FromClaims() error = nil, want missing issuer error")
	}
}

// TestExternalIdentityMergeFillsOptionalProfileClaims verifies profile enrichment.
func TestExternalIdentityMergeFillsOptionalProfileClaims(t *testing.T) {
	external := ExternalIdentity{Issuer: "issuer", Subject: "subject"}
	enriched := external.Merge(ExternalIdentity{
		Email:           "ian@example.test",
		EmailVerified:   true,
		DisplayName:     "Ian",
		PictureURL:      "https://example.test/avatar.png",
		PreferredLocale: "en",
		RawClaimsHash:   "profile-hash",
		Username:        "ian",
	})
	if enriched.Email != "ian@example.test" || enriched.DisplayName != "Ian" || !enriched.EmailVerified {
		t.Fatalf("Merge() = %+v, want profile claims", enriched)
	}
	if enriched.RawClaimsHash != "profile-hash" {
		t.Fatalf("RawClaimsHash = %q, want profile hash", enriched.RawClaimsHash)
	}
	if enriched.Issuer != "issuer" || enriched.Subject != "subject" {
		t.Fatalf("Merge() changed identity keys: %+v", enriched)
	}
}
