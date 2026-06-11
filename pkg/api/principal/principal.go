package principal

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ErrMissing reports that no authenticated principal exists.
var ErrMissing = errors.New("principal missing")

// LocalKey is the Fiber local key for principals.
const LocalKey = "realmkit.principal"

// Principal contains provider-neutral authenticated actor data.
type Principal struct {
	// UserID is the local RealmKit user identifier.
	UserID uuid.UUID `json:"user_id"`

	// Issuer is the trusted issuer.
	Issuer string `json:"issuer"`

	// Subject is the provider subject when available.
	Subject string `json:"subject,omitempty"`

	// SubjectHash is the log-safe subject hash.
	SubjectHash string `json:"subject_hash"`

	// Audience contains token audiences.
	Audience []string `json:"audience"`

	// Scopes contains token scopes.
	Scopes []string `json:"scopes"`

	// DevelopmentBypass reports whether development auth bypass created the principal.
	DevelopmentBypass bool `json:"development_bypass,omitempty"`
}

// WithContext returns a context carrying principal.
func WithContext(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, principal)
}

// FromContext returns the principal from ctx.
func FromContext(ctx context.Context) (Principal, bool) {
	value, ok := ctx.Value(contextKey{}).(Principal)
	return value, ok
}

// Set stores principal in Fiber locals and user context.
func Set(ctx *fiber.Ctx, value Principal) {
	ctx.Locals(LocalKey, value)
	ctx.SetUserContext(WithContext(ctx.UserContext(), value))
}

// Current returns the current Fiber principal.
func Current(ctx *fiber.Ctx) (Principal, bool) {
	value, ok := ctx.Locals(LocalKey).(Principal)
	return value, ok
}

// Require returns the current principal or ErrMissing.
func Require(ctx *fiber.Ctx) (Principal, error) {
	value, ok := Current(ctx)
	if !ok {
		return Principal{}, ErrMissing
	}
	return value, nil
}

// contextKey avoids collisions with external context keys.
type contextKey struct{}
