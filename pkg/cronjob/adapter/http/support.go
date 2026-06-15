package http

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/cronjob/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// writeJSON writes a JSON response.
func writeJSON(ctx *fiber.Ctx, status int, payload any) error {
	ctx.Set(headers.ContentType, "application/json")
	return ctx.Status(status).JSON(payload)
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

// requireCron verifies one cron job administration permission.
func requireCron(ctx *fiber.Ctx, checker groupsport.Checker, action groupsdomain.Action, jobKey string) error {
	target := httpguard.All(action, groupsdomain.ObjectCronJob)
	if strings.TrimSpace(jobKey) != "" {
		target = httpguard.Object(action, groupsdomain.ObjectCronJob, cronJobScopeID(jobKey))
	}
	_, err := httpguard.Require(ctx, checker, target)
	return err
}

// cronJobScopeID returns a stable UUID scope for one cron job key.
func cronJobScopeID(jobKey string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("realmkit:cronjob:"+jobKey))
}

// handleError maps cron errors to problem responses.
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
		return problem.Write(ctx, problem.New(fiber.StatusNotFound, "cronjob_not_found", "Cron job was not found."))
	case errors.Is(err, port.ErrPreconditionFailed):
		return problem.Write(
			ctx,
			problem.New(fiber.StatusPreconditionFailed, "cronjob_precondition_failed", "Cron job version did not match."),
		)
	case errors.Is(err, port.ErrNoDueJob):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "cronjob_not_due", "No cron job is due."))
	case errors.Is(err, port.ErrHandlerMissing):
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "cronjob_handler_missing", "Cron job handler is not registered."))
	default:
		return err
	}
}
