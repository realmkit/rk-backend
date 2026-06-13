package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/realmkit/rk-backend/pkg/identity"
)

// Token contains validated access token data.
type Token struct {
	// Identity is the provider-neutral external identity.
	Identity identity.ExternalIdentity

	// Audience contains token audiences.
	Audience []string

	// Scopes contains token scopes.
	Scopes []string
}

// Validator validates bearer access tokens.
type Validator struct {
	config Config
	keys   *keySet
}

// NewValidator creates an OIDC validator.
func NewValidator(config Config) *Validator {
	return &Validator{config: config, keys: newKeySet(config.IssuerURL)}
}

// Validate validates one bearer token.
func (validator *Validator) Validate(ctx context.Context, raw string) (Token, error) {
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
		jwt.WithIssuer(validator.config.IssuerURL),
		jwt.WithAudience(validator.config.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	parsed, err := parser.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return validator.keys.Key(ctx, token)
	})
	if err != nil {
		return Token{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !parsed.Valid {
		return Token{}, ErrInvalidToken
	}
	external, err := identity.FromClaims(map[string]any(claims))
	if err != nil {
		return Token{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return Token{Identity: external, Audience: audienceFromClaims(claims), Scopes: scopesFromClaims(claims)}, nil
}

// MergeIdentityToken enriches an API token with a verified OIDC ID token.
func (validator *Validator) MergeIdentityToken(ctx context.Context, token Token, raw string) (Token, error) {
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
		jwt.WithIssuer(validator.config.IssuerURL),
		jwt.WithAudience(validator.config.ClientID),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	parsed, err := parser.ParseWithClaims(raw, claims, func(jwtToken *jwt.Token) (any, error) {
		return validator.keys.Key(ctx, jwtToken)
	})
	if err != nil {
		return Token{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !parsed.Valid {
		return Token{}, ErrInvalidToken
	}
	external, err := identity.FromClaims(map[string]any(claims))
	if err != nil {
		return Token{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if token.Identity.Issuer != external.Issuer || token.Identity.Subject != external.Subject {
		return Token{}, ErrInvalidToken
	}
	token.Identity = token.Identity.Merge(external)
	return token, nil
}

// enrichIdentity merges standard OIDC userinfo claims when available.
func (validator *Validator) enrichIdentity(
	ctx context.Context,
	raw string,
	external identity.ExternalIdentity,
) identity.ExternalIdentity {
	userInfoClaims, err := validator.userInfoClaims(ctx, raw)
	if err != nil {
		return external
	}
	if _, ok := userInfoClaims["iss"]; !ok {
		userInfoClaims["iss"] = external.Issuer
	}
	enriched, err := identity.FromClaims(userInfoClaims)
	if err != nil {
		return external
	}
	return external.Merge(enriched)
}

// userInfoClaims fetches provider-owned profile claims from OIDC userinfo.
func (validator *Validator) userInfoClaims(ctx context.Context, raw string) (map[string]any, error) {
	endpoint, err := validator.userInfoEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+raw)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("userinfo fetch failed with status %d", response.StatusCode)
	}
	var claims map[string]any
	if err := json.NewDecoder(response.Body).Decode(&claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// userInfoEndpoint reads the provider userinfo endpoint from OIDC discovery.
func (validator *Validator) userInfoEndpoint(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		strings.TrimRight(validator.config.IssuerURL, "/")+"/.well-known/openid-configuration",
		nil,
	)
	if err != nil {
		return "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return "", fmt.Errorf("oidc discovery failed with status %d", response.StatusCode)
	}
	var discovery struct {
		UserInfoEndpoint string `json:"userinfo_endpoint"`
	}
	if err := json.NewDecoder(response.Body).Decode(&discovery); err != nil {
		return "", err
	}
	if strings.TrimSpace(discovery.UserInfoEndpoint) == "" {
		return "", fmt.Errorf("oidc discovery userinfo_endpoint is required")
	}
	return discovery.UserInfoEndpoint, nil
}

// audienceFromClaims returns audience claim values.
func audienceFromClaims(claims jwt.MapClaims) []string {
	switch value := claims["aud"].(type) {
	case string:
		return []string{value}
	case []any:
		items := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return items
	default:
		return nil
	}
}

// scopesFromClaims returns OAuth scopes.
func scopesFromClaims(claims jwt.MapClaims) []string {
	scope, _ := claims["scope"].(string)
	if strings.TrimSpace(scope) == "" {
		return nil
	}
	return strings.Fields(scope)
}
