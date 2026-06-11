package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
)

// definitionRequest contains metafield definition input.
type definitionRequest struct {
	OwnerType   domain.OwnerType `json:"owner_type"`
	Namespace   domain.Namespace `json:"namespace"`
	Key         domain.Key       `json:"key"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ValueType   domain.ValueType `json:"value_type"`
	List        bool             `json:"list"`
	Required    bool             `json:"required"`
	Rules       domain.Rules     `json:"rules"`
	SortOrder   int              `json:"sort_order"`
	Active      *bool            `json:"active"`
}

// definitionUpdateRequest contains mutable metafield definition input.
type definitionUpdateRequest struct {
	Name        *string       `json:"name"`
	Description *string       `json:"description"`
	Required    *bool         `json:"required"`
	Rules       *domain.Rules `json:"rules"`
	SortOrder   *int          `json:"sort_order"`
	Active      *bool         `json:"active"`
}

// definitionListResponse contains paginated definition output.
type definitionListResponse struct {
	Items         []domain.MetafieldDefinition `json:"items"`
	NextPageToken string                       `json:"next_page_token,omitempty"`
}

// createDefinition handles definition creation.
func (handler handler) createDefinition(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	definition, err := handler.services.Definitions.CreateDefinition(ctx.UserContext(), port.CreateDefinitionCommand{
		Definition: definitionFromRequest(request),
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusCreated, definition)
}

// listDefinitions handles definition listing.
func (handler handler) listDefinitions(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := port.DefinitionFilter{
		OwnerType: domain.OwnerType(ctx.Query("owner_type")),
		Namespace: domain.Namespace(ctx.Query("namespace")),
	}
	if ctx.Query("active") != "" {
		active := ctx.QueryBool("active")
		filter.Active = &active
	}
	result, err := handler.services.Definitions.ListDefinitions(ctx.UserContext(), port.ListDefinitionsQuery{Filter: filter, Page: page})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, definitionListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// getDefinition handles single definition reads.
func (handler handler) getDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	definition, err := handler.services.Definitions.GetDefinition(ctx.UserContext(), port.GetDefinitionQuery{ID: id})
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
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request definitionUpdateRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	current, err := handler.services.Definitions.GetDefinition(ctx.UserContext(), port.GetDefinitionQuery{ID: id})
	if err != nil {
		return handleError(ctx, err)
	}
	definition := applyDefinitionUpdate(current, request)
	updated, err := handler.services.Definitions.UpdateDefinition(
		ctx.UserContext(),
		port.UpdateDefinitionCommand{Definition: definition, ExpectedVersion: version},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// archiveDefinition handles definition archival.
func (handler handler) archiveDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.ArchiveDefinitionCommand{ID: id, ExpectedVersion: version}
	if err := handler.services.Definitions.ArchiveDefinition(ctx.UserContext(), command); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// definitionFromRequest maps request to domain.
func definitionFromRequest(request definitionRequest) domain.MetafieldDefinition {
	return domain.MetafieldDefinition{
		ID:          uuid.Nil,
		OwnerType:   request.OwnerType,
		Namespace:   request.Namespace,
		Key:         request.Key,
		Name:        request.Name,
		Description: request.Description,
		ValueType:   request.ValueType,
		List:        request.List,
		Required:    request.Required,
		Rules:       request.Rules,
		SortOrder:   request.SortOrder,
		Active:      activeFromPointer(request.Active),
		Version:     1,
	}
}

// applyDefinitionUpdate applies mutable fields.
func applyDefinitionUpdate(definition domain.MetafieldDefinition, request definitionUpdateRequest) domain.MetafieldDefinition {
	if request.Name != nil {
		definition.Name = *request.Name
	}
	if request.Description != nil {
		definition.Description = *request.Description
	}
	if request.Required != nil {
		definition.Required = *request.Required
	}
	if request.Rules != nil {
		definition.Rules = *request.Rules
	}
	if request.SortOrder != nil {
		definition.SortOrder = *request.SortOrder
	}
	if request.Active != nil {
		definition.Active = *request.Active
	}
	return definition
}
