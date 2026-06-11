package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

const currentUserIDHeader = "X-RealmKit-User-Id"

// decodeJSON decodes a strict JSON request body.
func decodeJSON(ctx *fiber.Ctx, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(ctx.Body())))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_json", "Request body must be valid JSON."),
		}
	}
	return nil
}

// writeJSON writes a JSON response payload.
func writeJSON(ctx *fiber.Ctx, status int, payload any) error {
	ctx.Set(headers.ContentType, "application/json")
	return ctx.Status(status).JSON(payload)
}

// handleError maps application errors to problem responses.
func handleError(ctx *fiber.Ctx, err error) error {
	var validation domain.ValidationError
	if errors.As(err, &validation) {
		payload := problem.New(fiber.StatusUnprocessableEntity, "validation_failed", "Request validation failed.")
		for _, violation := range validation.Violations {
			payload.Errors = append(payload.Errors, problem.FieldError{Field: violation.Field, Message: violation.Message})
		}
		return problem.Write(ctx, payload)
	}
	switch {
	case errors.Is(err, port.ErrNotFound):
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "punishment_not_found", "Punishment resource was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusPreconditionFailed, "punishment_precondition_failed", "Punishment version did not match."),
		)
	case errors.Is(err, port.ErrConflict):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "punishment_conflict", "Punishment conflicts with current state."))
	case errors.Is(err, port.ErrForbidden):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusForbidden, "punishment_restricted", "Action is restricted by an active punishment."),
		)
	default:
		return err
	}
}

// idFromParam parses a UUID path parameter.
func idFromParam(ctx *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params(name))
	if err != nil {
		return uuid.Nil, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_path_parameter", name+" must be a UUID."),
		}
	}
	return id, nil
}

// pageFromQuery parses cursor pagination query parameters.
func pageFromQuery(ctx *fiber.Ctx) (pagination.Page, error) {
	page, err := pagination.New(pagination.Request{Limit: ctx.QueryInt("page_size"), Cursor: ctx.Query("page_token")})
	if err != nil {
		return pagination.Page{}, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_pagination", "Pagination parameters are invalid."),
		}
	}
	return page, nil
}

// expectedVersion parses the required If-Match optimistic version.
func expectedVersion(ctx *fiber.Ctx) (uint64, error) {
	value := strings.Trim(ctx.Get(headers.IfMatch), `" `)
	if value == "" {
		return 0, problem.Error{
			Problem: problem.New(fiber.StatusPreconditionRequired, "if_match_required", "If-Match header is required."),
		}
	}
	version, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_if_match", "If-Match must contain a numeric version."),
		}
	}
	return version, nil
}

// requireIdempotency verifies retryable commands carry an idempotency key.
func requireIdempotency(ctx *fiber.Ctx) error {
	if strings.TrimSpace(ctx.Get(headers.IdempotencyKey)) != "" {
		return nil
	}
	return problem.Error{
		Problem: problem.New(fiber.StatusBadRequest, "idempotency_key_required", "Idempotency-Key header is required."),
	}
}

// currentUserID parses the current authenticated user header.
func currentUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	value := strings.TrimSpace(ctx.Get(currentUserIDHeader))
	if value == "" {
		return uuid.Nil, problem.Error{
			Problem: problem.New(fiber.StatusUnauthorized, "unauthenticated", currentUserIDHeader+" is required."),
		}
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_current_user", currentUserIDHeader+" must be a UUID."),
		}
	}
	return id, nil
}

// setETag writes an optimistic concurrency ETag.
func setETag(ctx *fiber.Ctx, version uint64) {
	ctx.Set(headers.ETag, `"`+strconv.FormatUint(version, 10)+`"`)
}
