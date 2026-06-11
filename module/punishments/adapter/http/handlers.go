package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	"github.com/niflaot/gamehub-go/module/punishments/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
)

type definitionRequest struct {
	Key                    domain.Key              `json:"key"`
	Name                   string                  `json:"name"`
	Description            string                  `json:"description"`
	Color                  domain.Color            `json:"color"`
	Severity               int                     `json:"severity"`
	Status                 domain.DefinitionStatus `json:"status"`
	DefaultDurationSeconds *int64                  `json:"default_duration_seconds"`
	MinDurationSeconds     *int64                  `json:"min_duration_seconds"`
	MaxDurationSeconds     *int64                  `json:"max_duration_seconds"`
	AllowPermanent         bool                    `json:"allow_permanent"`
	RequiresReason         bool                    `json:"requires_reason"`
	RequiresTargetIP       bool                    `json:"requires_target_ip"`
	DisplayOrder           int                     `json:"display_order"`
	Actions                []domain.ActionTemplate `json:"actions"`
}

func (handler handler) createDefinition(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	created, err := handler.services.Punishments.CreateDefinition(ctx.UserContext(), definitionFromRequest(request, uuid.Nil))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, created.Version)
	return writeJSON(ctx, fiber.StatusCreated, created)
}

func (handler handler) listDefinitions(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Punishments.ListDefinitions(
		ctx.UserContext(),
		port.DefinitionFilter{Status: domain.DefinitionStatus(ctx.Query("status"))},
		page,
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

func (handler handler) getDefinition(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	definition, err := handler.services.Punishments.GetDefinition(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusOK, definition)
}

func (handler handler) updateDefinition(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	updated, err := handler.services.Punishments.UpdateDefinition(ctx.UserContext(), definitionFromRequest(request, id), version)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

func (handler handler) deleteDefinition(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Punishments.DeleteDefinition(ctx.UserContext(), id, version); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (handler handler) reorderActions(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	var request struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	if err := handler.services.Punishments.ReorderDefinitionActions(ctx.UserContext(), id, request.IDs); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// definitionFromRequest maps HTTP definition input to domain state.
func definitionFromRequest(request definitionRequest, id uuid.UUID) domain.Definition {
	return domain.Definition{
		ID:                     id,
		Key:                    request.Key,
		Name:                   request.Name,
		Description:            request.Description,
		Color:                  request.Color,
		Severity:               request.Severity,
		Status:                 request.Status,
		DefaultDurationSeconds: request.DefaultDurationSeconds,
		MinDurationSeconds:     request.MinDurationSeconds,
		MaxDurationSeconds:     request.MaxDurationSeconds,
		AllowPermanent:         request.AllowPermanent,
		RequiresReason:         request.RequiresReason,
		RequiresTargetIP:       request.RequiresTargetIP,
		DisplayOrder:           request.DisplayOrder,
		Actions:                request.Actions,
	}
}

type issueRequest struct {
	DefinitionID  uuid.UUID         `json:"definition_id"`
	TargetUserID  uuid.UUID         `json:"target_user_id"`
	TargetIPHash  string            `json:"target_ip_hash"`
	IssuerType    domain.IssuerType `json:"issuer_type"`
	IssuerKey     string            `json:"issuer_key"`
	Reason        string            `json:"reason"`
	PrivateReason string            `json:"private_reason"`
	StartsAt      *time.Time        `json:"starts_at"`
	ExpiresAt     *time.Time        `json:"expires_at"`
	Source        string            `json:"source"`
}

// issueCommand maps HTTP issue input to an application command.
func issueCommand(ctx *fiber.Ctx, actor uuid.UUID, request issueRequest) port.IssueCommand {
	startsAt := time.Time{}
	if request.StartsAt != nil {
		startsAt = *request.StartsAt
	}
	issuerUserID := &actor
	return port.IssueCommand{
		ActorUserID:    actor,
		DefinitionID:   request.DefinitionID,
		TargetUserID:   request.TargetUserID,
		TargetIPHash:   request.TargetIPHash,
		IssuerType:     request.IssuerType,
		IssuerUserID:   issuerUserID,
		IssuerKey:      request.IssuerKey,
		Reason:         request.Reason,
		PrivateReason:  request.PrivateReason,
		StartsAt:       startsAt,
		ExpiresAt:      request.ExpiresAt,
		Source:         request.Source,
		IdempotencyKey: ctx.Get(headers.IdempotencyKey),
	}
}
