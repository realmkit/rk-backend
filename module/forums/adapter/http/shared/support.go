// Package shared contains transport helpers for forum HTTP route packages.
package shared

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/api/authgate"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// DecodeJSON decodes the request body into target.
func DecodeJSON(ctx *fiber.Ctx, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(ctx.Body())))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		payload := problem.New(fiber.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return problem.Error{Problem: payload}
	}
	return nil
}

// WriteJSON writes a JSON response.
func WriteJSON(ctx *fiber.Ctx, status int, payload any) error {
	ctx.Set(headers.ContentType, "application/json")
	return ctx.Status(status).JSON(payload)
}

// WriteNoContent writes a no content response.
func WriteNoContent(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
}

// HandleError maps application errors to problem responses.
func HandleError(ctx *fiber.Ctx, err error) error {
	var validation domain.ValidationError
	if errors.As(err, &validation) {
		return writeValidationError(ctx, validation)
	}
	switch {
	case errors.Is(err, port.ErrNotFound):
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "forum_not_found", "Forum resource was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(ctx, problem.New(fiber.StatusPreconditionFailed, "forum_precondition_failed", "Forum version did not match."))
	case errors.Is(err, port.ErrConflict):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "forum_conflict", "Forum resource conflicts with existing state."))
	case errors.Is(err, port.ErrForbidden):
		return problem.Write(ctx, problem.New(fiber.StatusForbidden, "permission_denied", "Permission was denied."))
	case errors.Is(err, port.ErrInvalidMove):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusConflict, "invalid_forum_move", "Forum cannot be moved to the requested location."),
		)
	default:
		return err
	}
}

// IDFromParam parses a UUID route parameter.
func IDFromParam(ctx *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params(name))
	if err != nil {
		payload := problem.New(fiber.StatusBadRequest, "invalid_path_parameter", name+" must be a UUID.")
		return uuid.Nil, problem.Error{Problem: payload}
	}
	return id, nil
}

// PageFromQuery parses pagination query parameters.
func PageFromQuery(ctx *fiber.Ctx) (pagination.Page, error) {
	request := pagination.Request{
		Limit:  ctx.QueryInt("page_size"),
		Cursor: ctx.Query("page_token"),
	}
	page, err := pagination.New(request)
	if err != nil {
		payload := problem.New(fiber.StatusBadRequest, "invalid_pagination", "Pagination parameters are invalid.")
		return pagination.Page{}, problem.Error{Problem: payload}
	}
	return page, nil
}

// ExpectedVersion returns the required If-Match version.
func ExpectedVersion(ctx *fiber.Ctx) (uint64, error) {
	value := strings.Trim(ctx.Get(headers.IfMatch), `" `)
	if value == "" {
		payload := problem.New(fiber.StatusPreconditionRequired, "if_match_required", "If-Match header is required.")
		return 0, problem.Error{Problem: payload}
	}
	version, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		payload := problem.New(fiber.StatusBadRequest, "invalid_if_match", "If-Match must contain a numeric version.")
		return 0, problem.Error{Problem: payload}
	}
	return version, nil
}

// RequireIdempotency verifies Idempotency-Key is present.
func RequireIdempotency(ctx *fiber.Ctx) error {
	if strings.TrimSpace(ctx.Get(headers.IdempotencyKey)) != "" {
		return nil
	}
	payload := problem.New(fiber.StatusBadRequest, "idempotency_key_required", "Idempotency-Key header is required.")
	return problem.Error{Problem: payload}
}

// SetETag writes a version ETag.
func SetETag(ctx *fiber.Ctx, version uint64) {
	ctx.Set(headers.ETag, `"`+strconv.FormatUint(version, 10)+`"`)
}

// CurrentUserID returns the required current user ID from the validated principal.
func CurrentUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	return authgate.RequireUserID(ctx)
}

// OptionalUserID returns a validated current user ID when present.
func OptionalUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	return authgate.OptionalUserID(ctx), nil
}

// writeValidationError writes domain validation errors.
func writeValidationError(ctx *fiber.Ctx, validation domain.ValidationError) error {
	payload := problem.New(fiber.StatusUnprocessableEntity, "validation_failed", "Request validation failed.")
	payload.Errors = make([]problem.FieldError, 0, len(validation.Violations))
	for _, violation := range validation.Violations {
		payload.Errors = append(payload.Errors, problem.FieldError{
			Field:   violation.Field,
			Message: violation.Message,
		})
	}
	return problem.Write(ctx, payload)
}
