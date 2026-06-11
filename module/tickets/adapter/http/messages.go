package http

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
)

// messageRequest is the ticket message DTO.
type messageRequest struct {
	Visibility          domain.MessageVisibility `json:"visibility"`
	ContentDocumentJSON json.RawMessage          `json:"content_document_json"`
	ContentText         string                   `json:"content_text"`
}

// evidenceRequest is the evidence write DTO.
type evidenceRequest struct {
	MessageID   *uuid.UUID               `json:"message_id"`
	AssetID     *uuid.UUID               `json:"asset_id"`
	ExternalURL string                   `json:"external_url"`
	Label       string                   `json:"label"`
	Description string                   `json:"description"`
	Visibility  domain.MessageVisibility `json:"visibility"`
}

// createMessage handles ticket replies.
func (handler handler) createMessage(ctx *fiber.Ctx) error {
	ticketID, actor, key, err := ticketActorAndKey(ctx)
	if err != nil {
		return err
	}
	var request messageRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	message, err := handler.services.Conversation.CreateMessage(ctx.Context(), port.MessageCommand{
		ActorUserID:         actor,
		TicketID:            ticketID,
		Visibility:          request.Visibility,
		ContentDocumentJSON: request.ContentDocumentJSON,
		ContentText:         request.ContentText,
		IdempotencyKey:      key,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, message.Version)
	return writeJSON(ctx, fiber.StatusCreated, message)
}

// listMessages handles ticket message reads.
func (handler handler) listMessages(ctx *fiber.Ctx) error {
	ticketID, err := idFromParam(ctx, "ticket_id")
	if err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	includeStaffOnly := ctx.QueryBool("include_staff_only")
	result, err := handler.services.Conversation.ListMessages(ctx.Context(), ticketID, actor, includeStaffOnly, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// addEvidence handles evidence additions.
func (handler handler) addEvidence(ctx *fiber.Ctx) error {
	ticketID, actor, key, err := ticketActorAndKey(ctx)
	if err != nil {
		return err
	}
	var request evidenceRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	evidence, err := handler.services.Conversation.AddEvidence(ctx.Context(), port.EvidenceCommand{
		ActorUserID:    actor,
		TicketID:       ticketID,
		MessageID:      request.MessageID,
		AssetID:        request.AssetID,
		ExternalURL:    request.ExternalURL,
		Label:          request.Label,
		Description:    request.Description,
		Visibility:     request.Visibility,
		IdempotencyKey: key,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusCreated, evidence)
}

// listEvidence handles evidence reads.
func (handler handler) listEvidence(ctx *fiber.Ctx) error {
	ticketID, err := idFromParam(ctx, "ticket_id")
	if err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	includeStaffOnly := ctx.QueryBool("include_staff_only")
	items, err := handler.services.Conversation.ListEvidence(ctx.Context(), ticketID, actor, includeStaffOnly)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, map[string]any{"items": items})
}

// ticketActorAndKey parses common mutating ticket route inputs.
func ticketActorAndKey(ctx *fiber.Ctx) (uuid.UUID, uuid.UUID, string, error) {
	ticketID, err := idFromParam(ctx, "ticket_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, "", err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", err
	}
	key, err := requireIdempotency(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", err
	}
	return ticketID, actor, key, nil
}
