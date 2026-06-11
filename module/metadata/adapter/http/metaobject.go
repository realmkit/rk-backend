package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
)

// metaobjectDefinitionRequest contains metaobject definition input.
type metaobjectDefinitionRequest struct {
	Type        domain.MetaobjectType    `json:"type"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Fields      []domain.FieldDefinition `json:"fields"`
	Active      *bool                    `json:"active"`
}

// metaobjectDefinitionUpdateRequest contains mutable metaobject definition input.
type metaobjectDefinitionUpdateRequest struct {
	Name        *string                   `json:"name"`
	Description *string                   `json:"description"`
	Fields      *[]domain.FieldDefinition `json:"fields"`
	Active      *bool                     `json:"active"`
}

// metaobjectDefinitionListResponse contains paginated metaobject definition output.
type metaobjectDefinitionListResponse struct {
	Items         []domain.MetaobjectDefinition `json:"items"`
	NextPageToken string                        `json:"next_page_token,omitempty"`
}

// createMetaobjectDefinition handles metaobject definition creation.
func (handler handler) createMetaobjectDefinition(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request metaobjectDefinitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	definition, err := handler.services.Metaobjects.CreateMetaobjectDefinition(ctx.UserContext(), port.CreateMetaobjectDefinitionCommand{
		Definition: metaobjectDefinitionFromRequest(request),
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusCreated, definition)
}

// listMetaobjectDefinitions handles metaobject definition listing.
func (handler handler) listMetaobjectDefinitions(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := port.MetaobjectDefinitionFilter{Type: domain.MetaobjectType(ctx.Query("type"))}
	if ctx.Query("active") != "" {
		active := ctx.QueryBool("active")
		filter.Active = &active
	}
	result, err := handler.services.Metaobjects.ListMetaobjectDefinitions(
		ctx.UserContext(),
		port.ListMetaobjectDefinitionsQuery{Filter: filter, Page: page},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, metaobjectDefinitionListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// getMetaobjectDefinition handles metaobject definition reads.
func (handler handler) getMetaobjectDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	definition, err := handler.services.Metaobjects.GetMetaobjectDefinition(ctx.UserContext(), port.GetMetaobjectDefinitionQuery{ID: id})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusOK, definition)
}

// updateMetaobjectDefinition handles metaobject definition updates.
func (handler handler) updateMetaobjectDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request metaobjectDefinitionUpdateRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	current, err := handler.services.Metaobjects.GetMetaobjectDefinition(ctx.UserContext(), port.GetMetaobjectDefinitionQuery{ID: id})
	if err != nil {
		return handleError(ctx, err)
	}
	updated, err := handler.services.Metaobjects.UpdateMetaobjectDefinition(ctx.UserContext(), port.UpdateMetaobjectDefinitionCommand{
		Definition:      applyMetaobjectDefinitionUpdate(current, request),
		ExpectedVersion: version,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// archiveMetaobjectDefinition handles metaobject definition archival.
func (handler handler) archiveMetaobjectDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.ArchiveMetaobjectDefinitionCommand{ID: id, ExpectedVersion: version}
	if err := handler.services.Metaobjects.ArchiveMetaobjectDefinition(ctx.UserContext(), command); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// metaobjectDefinitionFromRequest maps request to domain.
func metaobjectDefinitionFromRequest(request metaobjectDefinitionRequest) domain.MetaobjectDefinition {
	return domain.MetaobjectDefinition{
		Type:        request.Type,
		Name:        request.Name,
		Description: request.Description,
		Fields:      request.Fields,
		Active:      activeFromPointer(request.Active),
		Version:     1,
	}
}

// applyMetaobjectDefinitionUpdate applies mutable fields.
func applyMetaobjectDefinitionUpdate(
	definition domain.MetaobjectDefinition,
	request metaobjectDefinitionUpdateRequest,
) domain.MetaobjectDefinition {
	if request.Name != nil {
		definition.Name = *request.Name
	}
	if request.Description != nil {
		definition.Description = *request.Description
	}
	if request.Fields != nil {
		definition.Fields = *request.Fields
	}
	if request.Active != nil {
		definition.Active = *request.Active
	}
	return definition
}
