package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// writeJSON writes a JSON response.
func writeJSON(ctx *fiber.Ctx, status int, payload any) error {
	ctx.Set(headers.ContentType, "application/json")
	return ctx.Status(status).JSON(payload)
}

// idFromParam parses a UUID route parameter.
func idFromParam(ctx *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(ctx.Params(name))
	if err != nil {
		return uuid.Nil, problem.Error{Problem: problem.New(fiber.StatusBadRequest, "invalid_path_parameter", name+" must be a UUID.")}
	}
	return id, nil
}

// pageFromQuery parses pagination parameters.
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

// handleError maps event errors to problem responses.
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
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "event_not_found", "Event was not found."))
	case errors.Is(err, port.ErrForbidden):
		return problem.Write(ctx, problem.New(fiber.StatusForbidden, "event_forbidden", "Event scope is forbidden."))
	default:
		return err
	}
}
