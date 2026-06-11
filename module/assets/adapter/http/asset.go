package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/assets/domain"
	"github.com/niflaot/gamehub-go/module/assets/port"
)

// uploadIntentRequest is the create upload intent request body.
type uploadIntentRequest struct {
	Namespace       domain.Namespace   `json:"namespace"`
	Path            domain.VirtualPath `json:"path"`
	Filename        domain.Filename    `json:"filename"`
	DisplayName     string             `json:"display_name"`
	Visibility      domain.Visibility  `json:"visibility"`
	ContentType     string             `json:"content_type"`
	SizeBytes       int64              `json:"size_bytes"`
	CreatedByUserID *uuid.UUID         `json:"created_by_user_id"`
}

// updateAssetRequest is the update asset request body.
type updateAssetRequest struct {
	DisplayName string             `json:"display_name"`
	Path        domain.VirtualPath `json:"path"`
	Visibility  domain.Visibility  `json:"visibility"`
}

// uploadIntentResponse is the create upload intent response body.
type uploadIntentResponse struct {
	Asset     domain.Asset      `json:"asset"`
	UploadURL uploadURLResponse `json:"upload"`
}

// uploadURLResponse describes a signed upload request.
type uploadURLResponse struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expires_at"`
}

// assetListResponse contains one asset page.
type assetListResponse struct {
	Items         []domain.Asset `json:"items"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

// folderListResponse contains virtual folders.
type folderListResponse struct {
	Folders []string `json:"folders"`
}

// assetURLResponse contains a signed read URL.
type assetURLResponse struct {
	URL string `json:"url"`
}

// createUploadIntent creates an upload intent.
func (handler handler) createUploadIntent(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request uploadIntentRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	intent, err := handler.services.Assets.CreateUploadIntent(ctx.Context(), port.CreateUploadIntentCommand{
		Namespace:       request.Namespace,
		Path:            request.Path,
		Filename:        request.Filename,
		DisplayName:     request.DisplayName,
		Visibility:      request.Visibility,
		ContentType:     request.ContentType,
		SizeBytes:       request.SizeBytes,
		CreatedByUserID: request.CreatedByUserID,
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
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	asset, err := handler.services.Assets.CompleteUpload(ctx.Context(), port.CompleteUploadCommand{ID: id})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, asset.Version)
	return writeJSON(ctx, fiber.StatusOK, asset)
}

// getAsset returns one asset.
func (handler handler) getAsset(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	asset, err := handler.services.Assets.Get(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, asset.Version)
	return writeJSON(ctx, fiber.StatusOK, asset)
}

// getAssetURL returns a signed read URL.
func (handler handler) getAssetURL(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	url, err := handler.services.Assets.GetURL(ctx.Context(), id, 0)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, assetURLResponse{URL: url})
}

// listAssets lists assets.
func (handler handler) listAssets(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := assetFilterFromQuery(ctx)
	result, err := handler.services.Assets.List(ctx.Context(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, assetListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// listFolders lists virtual folders.
func (handler handler) listFolders(ctx *fiber.Ctx) error {
	folders, err := handler.services.Assets.ListFolders(ctx.Context(), port.FolderFilter{
		Namespace:  domain.Namespace(ctx.Query("namespace")),
		PathPrefix: domain.VirtualPath(ctx.Query("path_prefix")),
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, folderListResponse{Folders: folders})
}

// updateAsset updates mutable asset fields.
func (handler handler) updateAsset(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
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
		ctx.Context(),
		port.UpdateAssetCommand{
			ID:              id,
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
	id, err := idFromParam(ctx, "asset_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Assets.Delete(ctx.Context(), port.DeleteAssetCommand{ID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// assetFilterFromQuery returns asset filters from query params.
func assetFilterFromQuery(ctx *fiber.Ctx) port.AssetFilter {
	return port.AssetFilter{
		Namespace:  domain.Namespace(ctx.Query("namespace")),
		Path:       domain.VirtualPath(ctx.Query("path")),
		PathPrefix: domain.VirtualPath(ctx.Query("path_prefix")),
		Status:     domain.Status(ctx.Query("status")),
	}
}
