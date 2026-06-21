package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// uploadIntentRequest is the create upload intent request body.
type uploadIntentRequest struct {
	Namespace       domain.Namespace   `json:"namespace"`          // Namespace stores the namespace value.
	Path            domain.VirtualPath `json:"path"`               // Path stores the path value.
	Filename        domain.Filename    `json:"filename"`           // Filename stores the filename value.
	DisplayName     string             `json:"display_name"`       // DisplayName stores the display name value.
	Visibility      domain.Visibility  `json:"visibility"`         // Visibility stores the visibility value.
	ContentType     string             `json:"content_type"`       // ContentType stores the content type value.
	SizeBytes       int64              `json:"size_bytes"`         // SizeBytes stores the size bytes value.
	CreatedByUserID *uuid.UUID         `json:"created_by_user_id"` // CreatedByUserID stores the created by user i d value.
}

// updateAssetRequest is the update asset request body.
type updateAssetRequest struct {
	Namespace   domain.Namespace   `json:"namespace"`    // Namespace stores the namespace value.
	DisplayName string             `json:"display_name"` // DisplayName stores the display name value.
	Path        domain.VirtualPath `json:"path"`         // Path stores the path value.
	Visibility  domain.Visibility  `json:"visibility"`   // Visibility stores the visibility value.
}

// uploadIntentResponse is the create upload intent response body.
type uploadIntentResponse struct {
	Asset     domain.Asset      `json:"asset"`  // Asset stores the asset value.
	UploadURL uploadURLResponse `json:"upload"` // UploadURL stores the upload u r l value.
}

// uploadURLResponse describes a signed upload request.
type uploadURLResponse struct {
	Method    string            `json:"method"`     // Method stores the method value.
	URL       string            `json:"url"`        // URL stores the u r l value.
	Headers   map[string]string `json:"headers"`    // Headers stores the headers value.
	ExpiresAt time.Time         `json:"expires_at"` // ExpiresAt stores the expires at value.
}

// createUploadIntent creates an upload intent.
func (handler handler) createUploadIntent(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	if err := httpguard.Check(
		ctx,
		handler.services.Checker,
		actor,
		httpguard.All(groupsdomain.PermissionAssetsCreate, groupsdomain.ObjectAsset),
	); err != nil {
		return err
	}
	var request uploadIntentRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	intent, err := handler.services.Assets.CreateUploadIntent(ctx.UserContext(), port.CreateUploadIntentCommand{
		Namespace:       request.Namespace,
		Path:            request.Path,
		Filename:        request.Filename,
		DisplayName:     request.DisplayName,
		Visibility:      request.Visibility,
		ContentType:     request.ContentType,
		SizeBytes:       request.SizeBytes,
		CreatedByUserID: &actor,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, intent.Asset.Version)
	return writeJSON(
		ctx,
		fiber.StatusCreated,
		uploadIntentResponse{
			Asset:     intent.Asset,
			UploadURL: uploadURLResponse{Method: intent.Method, URL: intent.URL, Headers: intent.Headers, ExpiresAt: intent.ExpiresAt},
		},
	)
}

// completeUpload completes an upload intent.
func (handler handler) completeUpload(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(groupsdomain.PermissionAssetsUpdate, groupsdomain.ObjectAsset, id),
	); err != nil {
		return err
	}
	asset, err := handler.services.Assets.CompleteUpload(ctx.UserContext(), port.CompleteUploadCommand{ID: id})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, asset.Version)
	return writeJSON(ctx, fiber.StatusOK, asset)
}

// updateAsset updates mutable asset fields.
func (handler handler) updateAsset(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(groupsdomain.PermissionAssetsUpdate, groupsdomain.ObjectAsset, id),
	); err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request updateAssetRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	asset, err := handler.services.Assets.Update(
		ctx.UserContext(),
		port.UpdateAssetCommand{
			ID:              id,
			Namespace:       request.Namespace,
			DisplayName:     request.DisplayName,
			Path:            request.Path,
			Visibility:      request.Visibility,
			ExpectedVersion: version,
		},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, asset.Version)
	return writeJSON(ctx, fiber.StatusOK, asset)
}

// deleteAsset deletes an asset.
func (handler handler) deleteAsset(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(groupsdomain.PermissionAssetsDelete, groupsdomain.ObjectAsset, id),
	); err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Assets.Delete(ctx.UserContext(), port.DeleteAssetCommand{ID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}
