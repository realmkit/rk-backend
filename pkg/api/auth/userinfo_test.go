package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/realmkit/rk-backend/pkg/identity"
)

// TestValidatorEnrichesIdentityFromUserInfo verifies provider profile sync.
func TestValidatorEnrichesIdentityFromUserInfo(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	var issuer string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(writer).Encode(map[string]string{
				"jwks_uri":          issuer + "/jwks",
				"userinfo_endpoint": issuer + "/userinfo",
			})
		case "/jwks":
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"keys": []any{jwkFromKey("kid-1", &privateKey.PublicKey)},
			})
		case "/userinfo":
			if request.Header.Get("Authorization") == "" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"sub":            "subject",
				"email":          "ian@example.test",
				"email_verified": true,
				"name":           "Ian",
				"picture":        "https://example.test/ian.png",
			})
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	issuer = server.URL
	raw := signedUserInfoTestToken(t, issuer, privateKey)

	validator := NewValidator(Config{IssuerURL: issuer, Audience: "realmkit-api"})
	validated, err := validator.Validate(context.Background(), raw)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	validated.Identity = validator.enrichIdentity(context.Background(), raw, validated.Identity)
	if validated.Identity.Email != "ian@example.test" || validated.Identity.DisplayName != "Ian" {
		t.Fatalf("Identity = %+v, want userinfo profile claims", validated.Identity)
	}
	if validated.Identity.PictureURL != "https://example.test/ian.png" {
		t.Fatalf("PictureURL = %q, want userinfo picture", validated.Identity.PictureURL)
	}
}

// TestValidatorMergesVerifiedIDTokenClaims verifies browser profile token sync.
func TestValidatorMergesVerifiedIDTokenClaims(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	validator := NewValidator(Config{
		Audience:  "realmkit-api",
		ClientID:  "realmkit-frontend",
		IssuerURL: "https://auth.example.test/oidc",
	})
	validator.keys.keys["kid-1"] = &privateKey.PublicKey
	validator.keys.expires = time.Now().Add(time.Hour)

	accessToken, err := signedToken(
		privateKey,
		"realmkit-api",
		map[string]any{"iss": validator.config.IssuerURL, "sub": "subject"},
	)
	if err != nil {
		t.Fatalf("signedToken() access error = %v", err)
	}
	idToken, err := signedToken(
		privateKey,
		"realmkit-frontend",
		map[string]any{
			"email":          "ian@example.test",
			"email_verified": true,
			"iss":            validator.config.IssuerURL,
			"name":           "Ian",
			"picture":        "https://example.test/ian.png",
			"sub":            "subject",
		},
	)
	if err != nil {
		t.Fatalf("signedToken() id error = %v", err)
	}

	validated, err := validator.Validate(context.Background(), accessToken)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	enriched, err := validator.MergeIdentityToken(context.Background(), validated, idToken)
	if err != nil {
		t.Fatalf("MergeIdentityToken() error = %v", err)
	}
	if enriched.Identity.Email != "ian@example.test" || enriched.Identity.DisplayName != "Ian" {
		t.Fatalf("Identity = %+v, want id token profile claims", enriched.Identity)
	}
}

// TestValidatorRejectsMismatchedIDTokenSubject verifies subject binding.
func TestValidatorRejectsMismatchedIDTokenSubject(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	validator := NewValidator(Config{Audience: "api", ClientID: "client", IssuerURL: "issuer"})
	validator.keys.keys["kid-1"] = &privateKey.PublicKey
	validator.keys.expires = time.Now().Add(time.Hour)
	idToken, err := signedToken(privateKey, "client", map[string]any{"iss": "issuer", "sub": "other"})
	if err != nil {
		t.Fatalf("signedToken() error = %v", err)
	}
	_, err = validator.MergeIdentityToken(
		context.Background(),
		Token{Identity: identityWithSubject("issuer", "subject")},
		idToken,
	)
	if err == nil {
		t.Fatalf("MergeIdentityToken() error = nil, want mismatch rejection")
	}
}

// signedUserInfoTestToken signs a minimal API token for userinfo tests.
func signedUserInfoTestToken(t *testing.T, issuer string, privateKey *rsa.PrivateKey) string {
	t.Helper()
	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		jwt.MapClaims{
			"iss": issuer,
			"sub": "subject",
			"aud": "realmkit-api",
			"exp": time.Now().Add(time.Hour).Unix(),
		},
	)
	token.Header["kid"] = "kid-1"
	raw, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return raw
}

// identityWithSubject creates one minimal external identity.
func identityWithSubject(issuer string, subject string) identity.ExternalIdentity {
	return identity.ExternalIdentity{Issuer: issuer, Subject: subject}
}

// signedToken signs a token with common test claims.
func signedToken(privateKey *rsa.PrivateKey, audience string, claims map[string]any) (string, error) {
	tokenClaims := jwt.MapClaims{"aud": audience, "exp": time.Now().Add(time.Hour).Unix()}
	for key, value := range claims {
		tokenClaims[key] = value
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)
	token.Header["kid"] = "kid-1"
	return token.SignedString(privateKey)
}
