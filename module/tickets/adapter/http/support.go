package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/api/authgate"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
)

// decodeJSON decodes a strict JSON body.
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

// handleError maps service errors to problem responses.
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
	case errors.Is(err, search.ErrInvalidCursor):
		return problem.Write(ctx, problem.New(fiber.StatusBadRequest, "invalid_page_token", "Page token is invalid."))
	case errors.Is(err, port.ErrNotFound):
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "ticket_not_found", "Ticket resource was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusPreconditionFailed, "ticket_precondition_failed", "Ticket version did not match."),
		)
	case errors.Is(err, port.ErrConflict):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "ticket_conflict", "Ticket conflicts with current state."))
	case errors.Is(err, port.ErrForbidden):
		return problem.Write(ctx, problem.New(fiber.StatusForbidden, "ticket_forbidden", "Ticket action is not allowed."))
	default:
		return err
	}
}

// searchProblem maps invalid search parameters to a problem response.
func searchProblem(err error) error {
	code := "invalid_search"
	if errors.Is(err, search.ErrInvalidCursor) {
		code = "invalid_page_token"
	}
	return problem.Error{Problem: problem.New(fiber.StatusBadRequest, code, "Search parameters are invalid.")}
}

// idFromParam parses a UUID path parameter.
func idFromParam(ctx *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params(name))
	if err != nil {
		return uuid.Nil, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_path_parameter", name+" must be a UUID.")}
	}
	return id, nil
}

// pageFromQuery parses cursor pagination query parameters.
func pageFromQuery(ctx *fiber.Ctx) (pagination.Page, error) {
	page, err := pagination.New(pagination.Request{
		Limit:  ctx.QueryInt("page_size"),
		Cursor: ctx.Query("page_token"),
	})
	if err != nil {
		return pagination.Page{}, problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "invalid_pagination", "Pagination parameters are invalid."),
		}
	}
	return page, nil
}

// currentUserID returns the current authenticated user from the validated principal.
func currentUserID(ctx *fiber.Ctx) (uuid.UUID, error) {
	return authgate.RequireUserID(ctx)
}

// requireTicket verifies one ticket permission.
func requireTicket(ctx *fiber.Ctx, checker groupsport.Checker, action groupsdomain.Action, id uuid.UUID) error {
	target := httpguard.All(action, groupsdomain.ObjectTicket)
	if id != uuid.Nil {
		target = httpguard.Object(action, groupsdomain.ObjectTicket, id)
	}
	_, err := httpguard.Require(ctx, checker, target)
	return err
}

// expectedVersion parses required optimistic concurrency state.
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

// requireIdempotency returns the required idempotency key.
func requireIdempotency(ctx *fiber.Ctx) (string, error) {
	key := strings.TrimSpace(ctx.Get(headers.IdempotencyKey))
	if key == "" {
		return "", problem.Error{
			Problem: problem.New(fiber.StatusBadRequest, "idempotency_key_required", "Idempotency-Key header is required."),
		}
	}
	return key, nil
}

// setETag writes an optimistic concurrency ETag.
func setETag(ctx *fiber.Ctx, version uint64) {
	ctx.Set(headers.ETag, `"`+strconv.FormatUint(version, 10)+`"`)
}
