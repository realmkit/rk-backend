package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
)

// listEvents lists durable events.
func (handler handler) listEvents(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := port.ListFilter{
		Status:        domain.Status(ctx.Query("status")),
		Producer:      domain.Producer(ctx.Query("producer")),
		EventKey:      domain.EventKey(ctx.Query("event_key")),
		AggregateType: domain.AggregateType(ctx.Query("aggregate_type")),
	}
	events, err := handler.services.Events.List(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, events)
}

// getEvent returns one durable event.
func (handler handler) getEvent(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "event_id")
	if err != nil {
		return err
	}
	event, err := handler.services.Events.Get(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, event)
}

// replayEvent requeues one event.
func (handler handler) replayEvent(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "event_id")
	if err != nil {
		return err
	}
	if err := handler.services.Events.Replay(ctx.UserContext(), id); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// cancelEvent cancels one event.
func (handler handler) cancelEvent(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "event_id")
	if err != nil {
		return err
	}
	if err := handler.services.Events.Cancel(ctx.UserContext(), id); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
