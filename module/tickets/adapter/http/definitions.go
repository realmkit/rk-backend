package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/search"
)

// definitionRequest is the definition write DTO.
type definitionRequest struct {
	Key                     domain.Key              `json:"key"`                        // Key stores the key value.
	Name                    string                  `json:"name"`                       // Name stores the name value.
	Description             string                  `json:"description"`                // Description stores the description value.
	Kind                    domain.Kind             `json:"kind"`                       // Kind stores the kind value.
	Status                  domain.DefinitionStatus `json:"status"`                     // Status stores the status value.
	DefaultTeamGroupID      *uuid.UUID              `json:"default_team_group_id"`      // DefaultTeamGroupID stores the default team group i d value.
	DefaultAssigneeUserID   *uuid.UUID              `json:"default_assignee_user_id"`   // DefaultAssigneeUserID stores the default assignee user i d value.
	SubmitterCanClose       bool                    `json:"submitter_can_close"`        // SubmitterCanClose stores the submitter can close value.
	SubmitterCanReopen      bool                    `json:"submitter_can_reopen"`       // SubmitterCanReopen stores the submitter can reopen value.
	AllowAnonymousSubmitter bool                    `json:"allow_anonymous_submitter"`  // AllowAnonymousSubmitter stores the allow anonymous submitter value.
	RequiresTargetUser      bool                    `json:"requires_target_user"`       // RequiresTargetUser stores the requires target user value.
	RequiresPunishment      bool                    `json:"requires_punishment"`        // RequiresPunishment stores the requires punishment value.
	RequiresEvidence        bool                    `json:"requires_evidence"`          // RequiresEvidence stores the requires evidence value.
	MaxOpenPerSubmitter     int                     `json:"max_open_per_submitter"`     // MaxOpenPerSubmitter stores the max open per submitter value.
	ReopenWindowSeconds     int64                   `json:"reopen_window_seconds"`      // ReopenWindowSeconds stores the reopen window seconds value.
	SLAFirstResponseSeconds int64                   `json:"sla_first_response_seconds"` // SLAFirstResponseSeconds stores the s l a first response seconds value.
	SLAResolutionSeconds    int64                   `json:"sla_resolution_seconds"`     // SLAResolutionSeconds stores the s l a resolution seconds value.
	MetadataSchemaKey       string                  `json:"metadata_schema_key"`        // MetadataSchemaKey stores the metadata schema key value.
	DisplayOrder            int                     `json:"display_order"`              // DisplayOrder stores the display order value.
}

// createDefinition handles ticket definition creation.
func (handler handler) createDefinition(ctx *fiber.Ctx) error {
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManageDefinitions, uuid.Nil); err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	created, err := handler.services.Definitions.CreateDefinition(ctx.UserContext(), request.definition(uuid.Nil))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, created.Version)
	return writeJSON(ctx, fiber.StatusCreated, created)
}

// listDefinitions handles definition list reads.
func (handler handler) listDefinitions(ctx *fiber.Ctx) error {
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManageDefinitions, uuid.Nil); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := definitionFilter(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Definitions.ListDefinitions(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// definitionFilter maps query params into a definition filter.
func definitionFilter(ctx *fiber.Ctx) (port.DefinitionFilter, error) {
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.DefinitionFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(ctx.Query("sort"), ctx.Query("direction"), port.DefaultDefinitionSort(), port.AllowedDefinitionSorts())
	if err != nil {
		return port.DefinitionFilter{}, searchProblem(err)
	}
	return port.DefinitionFilter{
		Kind:   domain.Kind(ctx.Query("kind")),
		Status: domain.DefinitionStatus(ctx.Query("status")),
		Query:  query,
		Sort:   sort,
	}, nil
}

// getDefinition handles one definition read.
func (handler handler) getDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManageDefinitions, id); err != nil {
		return err
	}
	definition, err := handler.services.Definitions.GetDefinition(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusOK, definition)
}

// updateDefinition handles definition updates.
func (handler handler) updateDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManageDefinitions, id); err != nil {
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
	updated, err := handler.services.Definitions.UpdateDefinition(ctx.UserContext(), request.definition(id), version)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// deleteDefinition handles definition soft deletion.
func (handler handler) deleteDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManageDefinitions, id); err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Definitions.DeleteDefinition(ctx.UserContext(), id, version); err != nil {
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
