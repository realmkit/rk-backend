package http

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
)

// valueRequest contains an owner metafield value.
type valueRequest struct {
	Value json.RawMessage `json:"value"`
}

// ownerMetadataResponse contains owner metadata output.
type ownerMetadataResponse struct {
	Owner      port.OwnerRef             `json:"owner"`
	Metafields []port.OwnerMetafieldView `json:"metafields"`
}

// setValue handles owner value upserts.
func (handler handler) setValue(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	owner, err := ownerFromParams(ctx)
	if err != nil {
		return err
	}
	expected, err := optionalExpectedVersion(ctx)
	if err != nil {
		return err
	}
	var request valueRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	value, created, err := handler.services.Values.SetValue(ctx.UserContext(), port.SetValueCommand{
		Owner:           owner,
		Namespace:       domain.Namespace(ctx.Params("namespace")),
		Key:             domain.Key(ctx.Params("key")),
		RawValue:        request.Value,
		ExpectedVersion: expected,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, value.Version)
	if created {
		return writeJSON(ctx, fiber.StatusCreated, value)
	}
	return writeJSON(ctx, fiber.StatusOK, value)
}

// listValues handles owner metadata listing.
func (handler handler) listValues(ctx *fiber.Ctx) error {
	owner, err := ownerFromParams(ctx)
	if err != nil {
		return err
	}
	includeEmpty := true
	if ctx.Query("include_empty") != "" {
		includeEmpty = ctx.QueryBool("include_empty")
	}
	view, err := handler.services.Values.ListValuesForOwner(ctx.UserContext(), port.ListValuesForOwnerQuery{
		Owner:        owner,
		Namespace:    domain.Namespace(ctx.Query("namespace")),
		IncludeEmpty: includeEmpty,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, ownerMetadataResponse{Owner: view.Owner, Metafields: view.Metafields})
}

// getValue handles owner value reads.
func (handler handler) getValue(ctx *fiber.Ctx) error {
	owner, err := ownerFromParams(ctx)
	if err != nil {
		return err
	}
	value, err := handler.services.Values.GetValue(ctx.UserContext(), port.GetValueQuery{
		Owner:     owner,
		Namespace: domain.Namespace(ctx.Params("namespace")),
		Key:       domain.Key(ctx.Params("key")),
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, value.Version)
	return writeJSON(ctx, fiber.StatusOK, value)
}

// deleteValue handles owner value deletion.
func (handler handler) deleteValue(ctx *fiber.Ctx) error {
	owner, err := ownerFromParams(ctx)
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	err = handler.services.Values.DeleteValue(ctx.UserContext(), port.DeleteValueCommand{
		Owner:           owner,
		Namespace:       domain.Namespace(ctx.Params("namespace")),
		Key:             domain.Key(ctx.Params("key")),
		ExpectedVersion: version,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}
