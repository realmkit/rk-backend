package auth

import (
	"context"
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
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/principal"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/identity"
)

// TestConfigPublicSplitsScopes verifies public auth config mapping.
func TestConfigPublicSplitsScopes(t *testing.T) {
	public := Config{Provider: "logto", IssuerURL: "https://auth.example", Audience: "api", ClientID: "frontend", Scopes: "openid profile email"}.Public()
	if len(public.Scopes) != 3 || public.Scopes[0] != "openid" {
		t.Fatalf("Scopes = %v, want split scopes", public.Scopes)
	}
}

// TestMiddlewareDevelopmentBypass verifies development-only principal creation.
func TestMiddlewareDevelopmentBypass(t *testing.T) {
	userID := uuid.New()
	provisioner := &testProvisioner{principal: principal.Principal{UserID: userID, SubjectHash: "dev:" + userID.String(), DevelopmentBypass: true}}
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(Config{DevelopmentBypass: true}, nil, provisioner, MiddlewareConfig{Development: true}))
	app.Get("/", func(ctx *fiber.Ctx) error {
		current, err := principal.Require(ctx)
		if err != nil {
			t.Fatalf("Require() error = %v", err)
		}
		return ctx.JSON(current)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(DevUserIDHeader, userID.String())
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusOK || provisioner.developmentCalls != 1 {
		t.Fatalf("status=%d developmentCalls=%d, want success", res.StatusCode, provisioner.developmentCalls)
	}
}

// TestMiddlewareRejectsMissingToken verifies auth failures map to 401.
func TestMiddlewareRejectsMissingToken(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(Config{}, nil, &testProvisioner{}, MiddlewareConfig{}))
	app.Get("/", func(ctx *fiber.Ctx) error { return nil })
	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want 401", res.StatusCode)
	}
}

// TestMiddlewareRejectsInvalidBearer verifies bearer errors map to invalid token.
func TestMiddlewareRejectsInvalidBearer(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(Config{IssuerURL: "https://auth.example", Audience: "api"}, nil, &testProvisioner{}, MiddlewareConfig{}))
	app.Get("/", func(ctx *fiber.Ctx) error { return nil })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(headers.Authorization, "Bearer invalid")
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want 401", res.StatusCode)
	}
}

// TestMiddlewareMapsDisabledDevelopmentUser verifies disabled users map to forbidden.
func TestMiddlewareMapsDisabledDevelopmentUser(t *testing.T) {
	userID := uuid.New()
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(Config{DevelopmentBypass: true}, nil, &testProvisioner{err: ErrDisabledUser}, MiddlewareConfig{Development: true}))
	app.Get("/", func(ctx *fiber.Ctx) error { return nil })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(DevUserIDHeader, userID.String())
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusForbidden {
		t.Fatalf("StatusCode = %d, want 403", res.StatusCode)
	}
}

// TestValidatorValidatesRS256Token verifies OIDC discovery and JWKS validation.
func TestValidatorValidatesRS256Token(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	var issuer string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(writer).Encode(map[string]string{"jwks_uri": issuer + "/jwks"})
		case "/jwks":
			_ = json.NewEncoder(writer).Encode(map[string]any{"keys": []any{jwkFromKey("kid-1", &privateKey.PublicKey)}})
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	issuer = server.URL
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"iss": issuer, "sub": "subject", "aud": "gamehub-api", "exp": time.Now().Add(time.Hour).Unix(), "scope": "openid profile", "preferred_username": "ian"})
	token.Header["kid"] = "kid-1"
	raw, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	validated, err := NewValidator(Config{IssuerURL: issuer, Audience: "gamehub-api"}).Validate(context.Background(), raw)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.Identity.Subject != "subject" || len(validated.Scopes) != 2 {
		t.Fatalf("validated = %+v, want subject and scopes", validated)
	}
}

// TestRegisterServesPublicConfig verifies auth config route.
func TestRegisterServesPublicConfig(t *testing.T) {
	app := fiber.New()
	Register(app.Group("/api/v1"), Config{Provider: "logto", IssuerURL: "https://auth.example", Audience: "api", ClientID: "client", Scopes: "openid"})
	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/auth/config", nil))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", res.StatusCode)
	}
}

// TestClaimHelpersCoverAudienceAndScopeShapes verifies claim parser branches.
func TestClaimHelpersCoverAudienceAndScopeShapes(t *testing.T) {
	stringAudience := audienceFromClaims(jwt.MapClaims{"aud": "api"})
	arrayAudience := audienceFromClaims(jwt.MapClaims{"aud": []any{"api", "other"}})
	if len(stringAudience) != 1 || len(arrayAudience) != 2 {
		t.Fatalf("audiences = %v %v, want parsed values", stringAudience, arrayAudience)
	}
	if scopes := scopesFromClaims(jwt.MapClaims{"scope": " "}); len(scopes) != 0 {
		t.Fatalf("scopes = %v, want empty", scopes)
	}
}

// testProvisioner records auth provisioner calls.
type testProvisioner struct {
	principal        principal.Principal
	developmentCalls int
	err              error
}

// Provision returns the configured principal.
func (provisioner *testProvisioner) Provision(context.Context, identity.ExternalIdentity, Token) (principal.Principal, error) {
	return provisioner.principal, provisioner.err
}

// DevelopmentPrincipal returns the configured development principal.
func (provisioner *testProvisioner) DevelopmentPrincipal(context.Context, uuid.UUID) (principal.Principal, error) {
	provisioner.developmentCalls++
	return provisioner.principal, provisioner.err
}

// jwkFromKey maps an RSA public key into a JWK.
func jwkFromKey(kid string, key *rsa.PublicKey) map[string]string {
	return map[string]string{"kty": "RSA", "kid": kid, "alg": "RS256", "use": "sig", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}
}
