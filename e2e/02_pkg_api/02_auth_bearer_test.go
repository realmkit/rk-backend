package pkgapi_e2e

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/principal"
)

// TestAPIAuthBearerTokenFlow verifies production bearer validation through JWKS.
func TestAPIAuthBearerTokenFlow(t *testing.T) {
	issuer := newTestIssuer(t)
	userID := uuid.New()
	provisioner := &developmentProvisioner{
		principal: principal.Principal{
			UserID:      userID,
			SubjectHash: "bearer:" + userID.String(),
			Scopes:      []string{"openid", "admin:realmkit"},
		},
	}
	ecosystem := harness.New(t)
	ecosystem.App.Get(
		"/e2e/bearer",
		auth.Middleware(
			auth.Config{IssuerURL: issuer.URL, Audience: "realmkit-api"},
			nil,
			provisioner,
			auth.MiddlewareConfig{Log: ecosystem.Log},
		),
		func(ctx *fiber.Ctx) error {
			current, err := principal.Require(ctx)
			if err != nil {
				return err
			}
			return ctx.JSON(map[string]any{
				"user_id": current.UserID.String(),
				"scopes":  current.Scopes,
			})
		},
	)

	valid := bearerRequest(t, issuer.Token(t, tokenClaims{
		Issuer: issuer.URL, Audience: "realmkit-api", Subject: "subject", ExpiresAt: time.Now().Add(time.Hour),
	}))
	response := ecosystem.Test(t, valid)
	body := harness.ResponseBody(t, response)
	assertStatus(t, response.StatusCode, fiber.StatusOK, body)
	if !contains(body, userID.String()) {
		t.Fatalf("body = %q, want provisioned user", body)
	}

	assertBearerDenied(t, ecosystem, issuer.Token(t, tokenClaims{
		Issuer: issuer.URL, Audience: "other-api", Subject: "subject", ExpiresAt: time.Now().Add(time.Hour),
	}))
	assertBearerDenied(t, ecosystem, issuer.Token(t, tokenClaims{
		Issuer: issuer.URL + "/wrong", Audience: "realmkit-api", Subject: "subject", ExpiresAt: time.Now().Add(time.Hour),
	}))
	assertBearerDenied(t, ecosystem, issuer.Token(t, tokenClaims{
		Issuer: issuer.URL, Audience: "realmkit-api", Subject: "subject", ExpiresAt: time.Now().Add(-time.Hour),
	}))
}

// assertBearerDenied verifies invalid bearer tokens fail closed.
func assertBearerDenied(t *testing.T, ecosystem *harness.Ecosystem, token string) {
	t.Helper()
	response := ecosystem.Test(t, bearerRequest(t, token))
	body := harness.ResponseBody(t, response)
	assertStatus(t, response.StatusCode, fiber.StatusUnauthorized, body)
	assertProblemCode(t, body, "invalid_token")
}

// bearerRequest creates a bearer-protected e2e request.
func bearerRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	request := harness.JSONRequest(fiber.MethodGet, "/e2e/bearer", "")
	request.Header.Set(headers.Authorization, "Bearer "+token)
	return request
}

// tokenClaims contains minimal JWT claims for auth e2e tests.
type tokenClaims struct {
	Issuer    string
	Audience  string
	Subject   string
	ExpiresAt time.Time
}

// testIssuer is a local OIDC discovery and JWKS server.
type testIssuer struct {
	URL        string
	privateKey *rsa.PrivateKey
	server     *httptest.Server
}

// newTestIssuer starts a local issuer with one RSA signing key.
func newTestIssuer(t *testing.T) testIssuer {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	issuer := testIssuer{privateKey: privateKey}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(writer).Encode(map[string]string{"jwks_uri": issuer.URL + "/jwks"})
		case "/jwks":
			_ = json.NewEncoder(writer).Encode(map[string]any{"keys": []any{jwkFromPublicKey(&privateKey.PublicKey)}})
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	issuer.URL = server.URL
	issuer.server = server
	t.Cleanup(server.Close)
	return issuer
}

// Token signs one RS256 test access token.
func (issuer testIssuer) Token(t *testing.T, claims tokenClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":                claims.Issuer,
		"sub":                claims.Subject,
		"aud":                claims.Audience,
		"exp":                claims.ExpiresAt.Unix(),
		"scope":              "openid admin:realmkit",
		"preferred_username": "e2e",
	})
	token.Header["kid"] = "kid-1"
	raw, err := token.SignedString(issuer.privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return raw
}

// jwkFromPublicKey returns a public JWK for the test issuer.
func jwkFromPublicKey(key *rsa.PublicKey) map[string]string {
	return map[string]string{
		"kty": "RSA",
		"use": "sig",
		"kid": "kid-1",
		"alg": "RS256",
		"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
	}
}

// Compile-time check that developmentProvisioner supports auth provisioning.
var _ auth.Provisioner = (*developmentProvisioner)(nil)
