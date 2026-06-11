package http

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
)

// metaobjectEntryRequest contains metaobject entry input.
type metaobjectEntryRequest struct {
	Handle      domain.Handle                  `json:"handle"`
	DisplayName string                         `json:"display_name"`
	Fields      map[domain.Key]json.RawMessage `json:"fields"`
}

// metaobjectEntryUpdateRequest contains mutable metaobject entry input.
type metaobjectEntryUpdateRequest struct {
	DisplayName *string                        `json:"display_name"`
	Fields      map[domain.Key]json.RawMessage `json:"fields"`
}

// metaobjectEntryListResponse contains paginated metaobject entry output.
type metaobjectEntryListResponse struct {
	Items         []domain.MetaobjectEntry `json:"items"`
	NextPageToken string                   `json:"next_page_token,omitempty"`
}

// createMetaobjectEntry handles metaobject entry creation.
func (handler handler) createMetaobjectEntry(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	definitionID, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	var request metaobjectEntryRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	entry, err := handler.services.Metaobjects.CreateMetaobjectEntry(
		ctx.UserContext(),
		port.CreateMetaobjectEntryCommand{
			Entry: domain.MetaobjectEntry{
				DefinitionID: definitionID,
				Handle:       request.Handle,
				DisplayName:  request.DisplayName,
			},
			RawFields: request.Fields,
		},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, entry.Version)
	return writeJSON(ctx, fiber.StatusCreated, entry)
}

// listMetaobjectEntries handles entry listing.
func (handler handler) listMetaobjectEntries(ctx *fiber.Ctx) error {
	definitionID, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Metaobjects.ListMetaobjectEntries(
		ctx.UserContext(),
		port.ListMetaobjectEntriesQuery{DefinitionID: definitionID, Page: page},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, entryListResponse(result.Items, result.NextCursor))
}

// getMetaobjectEntry handles entry reads.
func (handler handler) getMetaobjectEntry(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "entry_id")
	if err != nil {
		return err
	}
	entry, err := handler.services.Metaobjects.GetMetaobjectEntry(
		ctx.UserContext(),
		port.GetMetaobjectEntryQuery{ID: id},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, entry.Version)
	return writeJSON(ctx, fiber.StatusOK, entry)
}

// updateMetaobjectEntry handles entry updates.
func (handler handler) updateMetaobjectEntry(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "entry_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request metaobjectEntryUpdateRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	current, err := handler.services.Metaobjects.GetMetaobjectEntry(
		ctx.UserContext(),
		port.GetMetaobjectEntryQuery{ID: id},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	if request.DisplayName != nil {
		current.DisplayName = *request.DisplayName
	}
	updated, err := handler.services.Metaobjects.UpdateMetaobjectEntry(
		ctx.UserContext(),
		port.UpdateMetaobjectEntryCommand{
			Entry:           current,
			RawFields:       request.Fields,
			ExpectedVersion: version,
		},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// deleteMetaobjectEntry handles entry deletion.
func (handler handler) deleteMetaobjectEntry(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "entry_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.DeleteMetaobjectEntryCommand{ID: id, ExpectedVersion: version}
	if err := handler.services.Metaobjects.DeleteMetaobjectEntry(ctx.UserContext(), command); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// entryListResponse creates a paginated entry response.
func entryListResponse(items []domain.MetaobjectEntry, cursor string) metaobjectEntryListResponse {
	return metaobjectEntryListResponse{Items: items, NextPageToken: cursor}
}
