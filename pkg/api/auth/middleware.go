package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/identity"
	"go.uber.org/zap"
)

// Provisioner resolves validated identities into RealmKit principals.
type Provisioner interface {
	// Provision resolves or creates the local user for identity.
	Provision(ctx context.Context, external identity.ExternalIdentity, token Token) (principal.Principal, error)

	// DevelopmentPrincipal returns a principal for an existing local user.
	DevelopmentPrincipal(ctx context.Context, userID uuid.UUID) (principal.Principal, error)
}

// MiddlewareConfig configures auth middleware behavior.
type MiddlewareConfig struct {
	// Development reports whether the runtime is development.
	Development bool

	// Log receives development bypass warnings.
	Log *zap.Logger
}

// Middleware validates bearer tokens or development bypass headers.
func Middleware(config Config, validator *Validator, provisioner Provisioner, settings MiddlewareConfig) fiber.Handler {
	if validator == nil {
		validator = NewValidator(config)
	}
	if settings.Log == nil {
		settings.Log = zap.NewNop()
	}
	return func(ctx *fiber.Ctx) error {
		if token := bearerToken(ctx.Get(headers.Authorization)); token != "" {
			return authenticateBearer(ctx, validator, provisioner, settings.Log, token)
		}
		if config.DevelopmentBypass && settings.Development {
			return authenticateDevelopment(ctx, provisioner, settings.Log)
		}
		return problem.Write(ctx, problem.New(fiber.StatusUnauthorized, "unauthenticated", "Authorization bearer token is required."))
	}
}

// OptionalMiddleware validates bearer tokens when present and allows anonymous requests.
func OptionalMiddleware(config Config, validator *Validator, provisioner Provisioner, settings MiddlewareConfig) fiber.Handler {
	if validator == nil {
		validator = NewValidator(config)
	}
	if settings.Log == nil {
		settings.Log = zap.NewNop()
	}
	return func(ctx *fiber.Ctx) error {
		if token := bearerToken(ctx.Get(headers.Authorization)); token != "" {
			return authenticateBearer(ctx, validator, provisioner, settings.Log, token)
		}
		if config.DevelopmentBypass && settings.Development && strings.TrimSpace(ctx.Get(DevUserIDHeader)) != "" {
			return authenticateDevelopment(ctx, provisioner, settings.Log)
		}
		return ctx.Next()
	}
}

// Register registers public auth routes.
func Register(router fiber.Router, config Config) {
	router.Get("/auth/config", func(ctx *fiber.Ctx) error {
		ctx.Set(headers.ContentType, "application/json")
		return ctx.Status(fiber.StatusOK).JSON(config.Public())
	})
}

// authenticateBearer validates bearer token and stores a principal.
func authenticateBearer(
	ctx *fiber.Ctx,
	validator *Validator,
	provisioner Provisioner,
	log *zap.Logger,
	raw string,
) error {
	token, err := validator.Validate(ctx.UserContext(), raw)
	if err != nil {
		log.Warn("bearer token rejected", zap.Error(err))
		return authProblem(ctx, err)
	}
	if identityToken := strings.TrimSpace(ctx.Get(headers.IdentityToken)); identityToken != "" {
		token, err = validator.MergeIdentityToken(ctx.UserContext(), token, identityToken)
		if err != nil {
			log.Warn("identity token rejected", zap.Error(err))
			return authProblem(ctx, err)
		}
	} else {
		token.Identity = validator.enrichIdentity(ctx.UserContext(), raw, token.Identity)
	}
	current, err := provisioner.Provision(ctx.UserContext(), token.Identity, token)
	if err != nil {
		log.Warn("bearer principal provisioning failed", zap.Error(err))
		return authProblem(ctx, err)
	}
	principal.Set(ctx, current)
	return ctx.Next()
}

// authenticateDevelopment authenticates through the development-only header.
func authenticateDevelopment(ctx *fiber.Ctx, provisioner Provisioner, log *zap.Logger) error {
	value := strings.TrimSpace(ctx.Get(DevUserIDHeader))
	if value == "" {
		return problem.Write(ctx, problem.New(fiber.StatusUnauthorized, "unauthenticated", "Authorization bearer token is required."))
	}
	userID, err := uuid.Parse(value)
	if err != nil {
		return problem.Write(ctx, problem.New(fiber.StatusBadRequest, "invalid_development_user", DevUserIDHeader+" must be a UUID."))
	}
	current, err := provisioner.DevelopmentPrincipal(ctx.UserContext(), userID)
	if err != nil {
		return authProblem(ctx, err)
	}
	log.Warn("development auth bypass used", zap.String("user_id", userID.String()))
	principal.Set(ctx, current)
	return ctx.Next()
}

// bearerToken returns an authorization bearer token.
func bearerToken(value string) string {
	prefix := "Bearer "
	if len(value) < len(prefix) || !strings.EqualFold(value[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(value[len(prefix):])
}

// authProblem maps auth errors to problem responses.
func authProblem(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrDisabledUser):
		return problem.Write(ctx, problem.New(fiber.StatusForbidden, "user_disabled", "User is disabled."))
	case errors.Is(err, ErrInvalidToken):
		return problem.Write(ctx, problem.New(fiber.StatusUnauthorized, "invalid_token", "Bearer token is invalid."))
	default:
		return err
	}
}
