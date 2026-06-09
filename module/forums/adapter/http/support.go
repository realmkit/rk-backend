package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// currentUserIDHeader is the temporary debug current user header.
const currentUserIDHeader = "X-GameHub-User-Id"

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

// writeNoContent writes a no content response.
func writeNoContent(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
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
	case errors.Is(err, port.ErrNotFound):
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "forum_not_found", "Forum resource was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(ctx, problem.New(fiber.StatusPreconditionFailed, "forum_precondition_failed", "Forum version did not match."))
	case errors.Is(err, port.ErrConflict):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "forum_conflict", "Forum resource conflicts with existing state."))
	case errors.Is(err, port.ErrForbidden):
		return problem.Write(ctx, problem.New(fiber.StatusForbidden, "permission_denied", "Permission was denied."))
	case errors.Is(err, port.ErrInvalidMove):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "invalid_forum_move", "Forum cannot be moved to the requested location."))
	default:
		return err
	}
}

// idFromParam parses a UUID route parameter.
func idFromParam(ctx *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params(name))
	if err != nil {
		return uuid.Nil, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_path_parameter", name+" must be a UUID.")}
	}
	return id, nil
}

// pageFromQuery parses pagination query parameters.
func pageFromQuery(ctx *fiber.Ctx) (pagination.Page, error) {
	page, err := pagination.New(pagination.Request{Limit: ctx.QueryInt("page_size"), Cursor: ctx.Query("page_token")})
	if err != nil {
		return pagination.Page{}, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_pagination", "Pagination parameters are invalid.")}
	}
	return page, nil
}

// expectedVersion returns the required If-Match version.
func expectedVersion(ctx *fiber.Ctx) (uint64, error) {
	value := strings.Trim(ctx.Get(headers.IfMatch), `" `)
	if value == "" {
		return 0, problem.Error{Problem: problem.New(fiber.StatusPreconditionRequired, "if_match_required", "If-Match header is required.")}
	}
	version, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_if_match", "If-Match must contain a numeric version.")}
	}
	return version, nil
}

// requireIdempotency verifies Idempotency-Key is present.
func requireIdempotency(ctx *fiber.Ctx) error {
	if strings.TrimSpace(ctx.Get(headers.IdempotencyKey)) == "" {
		return problem.Error{Problem: problem.New(fiber.StatusBadRequest, "idempotency_key_required", "Idempotency-Key header is required.")}
	}
	return nil
}

// setETag writes a version ETag.
func setETag(ctx *fiber.Ctx, version uint64) {
	ctx.Set(headers.ETag, `"`+strconv.FormatUint(version, 10)+`"`)
}

// currentUserID returns the required current user ID from the temporary debug header.
func currentUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	id, err := optionalUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	if id == uuid.Nil {
		return uuid.Nil, problem.Error{Problem: problem.New(fiber.StatusUnauthorized, "unauthenticated", currentUserIDHeader+" is required.")}
	}
	return id, nil
}

// optionalUserID returns a current user ID when present.
func optionalUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	value := strings.TrimSpace(ctx.Get(currentUserIDHeader))
	if value == "" {
		return uuid.Nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_current_user", currentUserIDHeader+" must be a UUID.")}
	}
	return id, nil
}
