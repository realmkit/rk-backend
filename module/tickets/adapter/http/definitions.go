package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
)

// definitionRequest is the definition write DTO.
type definitionRequest struct {
	Key                     domain.Key              `json:"key"`
	Name                    string                  `json:"name"`
	Description             string                  `json:"description"`
	Kind                    domain.Kind             `json:"kind"`
	Status                  domain.DefinitionStatus `json:"status"`
	DefaultTeamGroupID      *uuid.UUID              `json:"default_team_group_id"`
	DefaultAssigneeUserID   *uuid.UUID              `json:"default_assignee_user_id"`
	SubmitterCanClose       bool                    `json:"submitter_can_close"`
	SubmitterCanReopen      bool                    `json:"submitter_can_reopen"`
	AllowAnonymousSubmitter bool                    `json:"allow_anonymous_submitter"`
	RequiresTargetUser      bool                    `json:"requires_target_user"`
	RequiresPunishment      bool                    `json:"requires_punishment"`
	RequiresEvidence        bool                    `json:"requires_evidence"`
	MaxOpenPerSubmitter     int                     `json:"max_open_per_submitter"`
	ReopenWindowSeconds     int64                   `json:"reopen_window_seconds"`
	SLAFirstResponseSeconds int64                   `json:"sla_first_response_seconds"`
	SLAResolutionSeconds    int64                   `json:"sla_resolution_seconds"`
	MetadataSchemaKey       string                  `json:"metadata_schema_key"`
	DisplayOrder            int                     `json:"display_order"`
}

// createDefinition handles ticket definition creation.
func (handler handler) createDefinition(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	created, err := handler.services.Definitions.CreateDefinition(ctx.Context(), request.definition(uuid.Nil))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, created.Version)
	return writeJSON(ctx, fiber.StatusCreated, created)
}

// listDefinitions handles definition list reads.
func (handler handler) listDefinitions(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := port.DefinitionFilter{
		Kind:   domain.Kind(ctx.Query("kind")),
		Status: domain.DefinitionStatus(ctx.Query("status")),
	}
	result, err := handler.services.Definitions.ListDefinitions(ctx.Context(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// getDefinition handles one definition read.
func (handler handler) getDefinition(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	definition, err := handler.services.Definitions.GetDefinition(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusOK, definition)
}

// updateDefinition handles definition updates.
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
	updated, err := handler.services.Definitions.UpdateDefinition(ctx.Context(), request.definition(id), version)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// deleteDefinition handles definition soft deletion.
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
	if err := handler.services.Definitions.DeleteDefinition(ctx.Context(), id, version); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// definition maps the request into domain state.
func (request definitionRequest) definition(id uuid.UUID) domain.Definition {
	if id == uuid.Nil {
		id = uuid.New()
	}
	return domain.Definition{
		ID:                      id,
		Key:                     request.Key,
		Name:                    request.Name,
		Description:             request.Description,
		Kind:                    request.Kind,
		Status:                  request.Status,
		DefaultTeamGroupID:      request.DefaultTeamGroupID,
		DefaultAssigneeUserID:   request.DefaultAssigneeUserID,
		SubmitterCanClose:       request.SubmitterCanClose,
		SubmitterCanReopen:      request.SubmitterCanReopen,
		AllowAnonymousSubmitter: request.AllowAnonymousSubmitter,
		RequiresTargetUser:      request.RequiresTargetUser,
		RequiresPunishment:      request.RequiresPunishment,
		RequiresEvidence:        request.RequiresEvidence,
		MaxOpenPerSubmitter:     request.MaxOpenPerSubmitter,
		ReopenWindowSeconds:     request.ReopenWindowSeconds,
		SLAFirstResponseSeconds: request.SLAFirstResponseSeconds,
		SLAResolutionSeconds:    request.SLAResolutionSeconds,
		MetadataSchemaKey:       request.MetadataSchemaKey,
		DisplayOrder:            request.DisplayOrder,
		Version:                 1,
	}
}
