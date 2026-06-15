package http

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/search"
)

// createTicketRequest is the ticket intake DTO.
type createTicketRequest struct {
	DefinitionID        uuid.UUID       `json:"definition_id"`
	Title               string          `json:"title"`
	SubmitterUserID     *uuid.UUID      `json:"submitter_user_id"`
	TargetUserID        *uuid.UUID      `json:"target_user_id"`
	PunishmentID        *uuid.UUID      `json:"punishment_id"`
	ContentDocumentJSON json.RawMessage `json:"content_document_json"`
	ContentText         string          `json:"content_text"`
	EvidenceAssetIDs    []uuid.UUID     `json:"evidence_asset_ids"`
}

// createTicket handles ticket intake.
func (handler handler) createTicket(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	key, err := requireIdempotency(ctx)
	if err != nil {
		return err
	}
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsCreate, uuid.Nil); err != nil {
		return err
	}
	var request createTicketRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	ticket, err := handler.services.Tickets.CreateTicket(ctx.UserContext(), request.command(actor, key))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, ticket.Version)
	return writeJSON(ctx, fiber.StatusCreated, ticket)
}

// createAppeal handles punishment appeal intake.
func (handler handler) createAppeal(ctx *fiber.Ctx) error {
	punishmentID, err := idFromParam(ctx, "punishment_id")
	if err != nil {
		return err
	}
	var request createTicketRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	request.PunishmentID = &punishmentID
	return handler.createTicketWithRequest(ctx, request)
}

// createTicketWithRequest creates a ticket from a prepared request.
func (handler handler) createTicketWithRequest(ctx *fiber.Ctx, request createTicketRequest) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	key, err := requireIdempotency(ctx)
	if err != nil {
		return err
	}
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsCreate, uuid.Nil); err != nil {
		return err
	}
	ticket, err := handler.services.Tickets.CreateTicket(ctx.UserContext(), request.command(actor, key))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, ticket.Version)
	return writeJSON(ctx, fiber.StatusCreated, ticket)
}

// listTickets handles queue and personal ticket reads.
func (handler handler) listTickets(ctx *fiber.Ctx) error {
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsView, uuid.Nil); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := ticketFilter(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Tickets.ListTickets(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// getTicket handles one ticket read.
func (handler handler) getTicket(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "ticket_id")
	if err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	ticket, err := handler.services.Tickets.GetTicket(ctx.UserContext(), id, actor)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, ticket.Version)
	return writeJSON(ctx, fiber.StatusOK, ticket)
}

// command maps intake request into an application command.
func (request createTicketRequest) command(actor uuid.UUID, key string) port.CreateTicketCommand {
	submitter := request.SubmitterUserID
	if submitter == nil {
		submitter = &actor
	}
	return port.CreateTicketCommand{
		ActorUserID:         actor,
		DefinitionID:        request.DefinitionID,
		Title:               request.Title,
		SubmitterUserID:     submitter,
		TargetUserID:        request.TargetUserID,
		PunishmentID:        request.PunishmentID,
		ContentDocumentJSON: request.ContentDocumentJSON,
		ContentText:         request.ContentText,
		EvidenceAssetIDs:    request.EvidenceAssetIDs,
		IdempotencyKey:      key,
	}
}

// ticketFilter parses queue filters.
func ticketFilter(ctx *fiber.Ctx) (port.TicketFilter, error) {
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.TicketFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(ctx.Query("sort"), ctx.Query("direction"), port.DefaultTicketSort(), port.AllowedTicketSorts())
	if err != nil {
		return port.TicketFilter{}, searchProblem(err)
	}
	return port.TicketFilter{
		SubmitterUserID:    queryUUID(ctx, "submitter_user_id"),
		TargetUserID:       queryUUID(ctx, "target_user_id"),
		PunishmentID:       queryUUID(ctx, "punishment_id"),
		CurrentTeamGroupID: queryUUID(ctx, "current_team_group_id"),
		AssigneeUserID:     queryUUID(ctx, "assignee_user_id"),
		Status:             domain.TicketStatus(ctx.Query("status")),
		Kind:               domain.Kind(ctx.Query("kind")),
		Query:              query,
		Sort:               sort,
	}, nil
}

// queryUUID parses optional UUID query parameters.
func queryUUID(ctx *fiber.Ctx, name string) uuid.UUID {
	value := ctx.Query(name)
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return id
}
