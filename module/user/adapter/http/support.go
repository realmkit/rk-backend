package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/user/domain"
	"github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/principal"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
)

// errAccountURLUnavailable reports that provider account URL is not available.
var errAccountURLUnavailable = errors.New("provider account url unavailable")

// decodeJSON decodes the request body into target.
func decodeJSON(ctx *fiber.Ctx, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(ctx.Body())))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")}
	}
	return nil
}

// writeJSON writes a JSON response.
func writeJSON(ctx *fiber.Ctx, status int, payload any) error {
	ctx.Set(headers.ContentType, "application/json")
	return ctx.Status(status).JSON(payload)
}

// handleError maps application errors to problem responses.
func handleError(ctx *fiber.Ctx, err error) error {
	var validation domain.ValidationError
	if errors.As(err, &validation) {
		payload := problem.New(fiber.StatusUnprocessableEntity, "validation_failed", "Request validation failed.")
		payload.Errors = make([]problem.FieldError, 0, len(validation.Violations))
		for _, violation := range validation.Violations {
			payload.Errors = append(payload.Errors, problem.FieldError{Field: violation.Field, Message: violation.Message})
		}
		return problem.Write(ctx, payload)
	}
	switch {
	case errors.Is(err, principal.ErrMissing):
		return problem.Write(ctx, problem.New(fiber.StatusUnauthorized, "unauthenticated", "Authentication is required."))
	case errors.Is(err, port.ErrNotFound):
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "user_not_found", "User was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(ctx, problem.New(fiber.StatusPreconditionFailed, "user_precondition_failed", "User version did not match."))
	case errors.Is(err, port.ErrDisabled):
		return problem.Write(ctx, problem.New(fiber.StatusForbidden, "user_disabled", "User is disabled."))
	case errors.Is(err, errAccountURLUnavailable):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusNotFound, "provider_account_url_unavailable", "Provider account URL is not configured."),
		)
	default:
		return err
	}
}

// expectedVersion returns the required If-Match version.
func expectedVersion(ctx *fiber.Ctx) (uint64, error) {
	value := strings.Trim(ctx.Get(headers.IfMatch), `" `)
	if value == "" {
		return 0, problem.Error{Problem: problem.New(fiber.StatusPreconditionRequired, "if_match_required", "If-Match header is required.")}
	}
	version, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_if_match", "If-Match must contain a numeric version."),
		}
	}
	return version, nil
}

// requireIdempotency verifies Idempotency-Key is present.
func requireIdempotency(ctx *fiber.Ctx) error {
	if strings.TrimSpace(ctx.Get(headers.IdempotencyKey)) == "" {
		return problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "idempotency_key_required", "Idempotency-Key header is required."),
		}
	}
	return nil
}

// setETag writes a version ETag.
func setETag(ctx *fiber.Ctx, version uint64) {
	ctx.Set(headers.ETag, `"`+strconv.FormatUint(version, 10)+`"`)
}
