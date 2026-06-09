package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/niflaot/gamehub-go/pkg/identity"
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
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
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
