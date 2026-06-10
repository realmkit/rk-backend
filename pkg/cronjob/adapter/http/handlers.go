package http

import "github.com/gofiber/fiber/v2"

// listDefinitions lists cron jobs.
func (handler handler) listDefinitions(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	definitions, err := handler.services.Cron.ListDefinitions(ctx.UserContext(), page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, definitions)
}

// getDefinition returns one cron job.
func (handler handler) getDefinition(ctx *fiber.Ctx) error {
	definition, err := handler.services.Cron.GetDefinition(ctx.UserContext(), ctx.Params("job_key"))
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, definition)
}

// listRuns lists run history.
func (handler handler) listRuns(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	runs, err := handler.services.Cron.ListRuns(ctx.UserContext(), ctx.Params("job_key"), page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, runs)
}

// runNow manually runs one job.
func (handler handler) runNow(ctx *fiber.Ctx) error {
	result, err := handler.services.Cron.Trigger(ctx.UserContext(), ctx.Params("job_key"))
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// pause disables one job.
func (handler handler) pause(ctx *fiber.Ctx) error {
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Cron.Pause(ctx.UserContext(), ctx.Params("job_key"), version); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// resume enables one job.
func (handler handler) resume(ctx *fiber.Ctx) error {
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Cron.Resume(ctx.UserContext(), ctx.Params("job_key"), version); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// repairLocks repairs stale locks.
func (handler handler) repairLocks(ctx *fiber.Ctx) error {
	count, err := handler.services.Cron.RepairLocks(ctx.UserContext())
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, map[string]any{"repaired": count})
}
