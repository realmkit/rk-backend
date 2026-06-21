package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// assetListResponse contains one asset page.
type assetListResponse struct {
	Items         []domain.Asset `json:"items"`                     // Items stores the items value.
	NextPageToken string         `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
	Query         string         `json:"query,omitempty"`           // Query stores the query value.
	Sort          string         `json:"sort,omitempty"`            // Sort stores the sort value.
	Direction     string         `json:"direction,omitempty"`       // Direction stores the direction value.
}

// folderListResponse contains virtual folders.
type folderListResponse struct {
	Folders []string `json:"folders"` // Folders stores the folders value.
}

// namespaceListResponse contains asset namespaces.
type namespaceListResponse struct {
	Namespaces []string `json:"namespaces"` // Namespaces stores the namespaces value.
}

// assetURLResponse contains a signed read URL.
type assetURLResponse struct {
	URL string `json:"url"` // URL stores the u r l value.
}

// getAsset returns one asset.
func (handler handler) getAsset(ctx *fiber.Ctx) error {
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
		httpguard.Object(groupsdomain.PermissionAssetsView, groupsdomain.ObjectAsset, id),
	); err != nil {
		return err
	}
	asset, err := handler.services.Assets.Get(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, asset.Version)
	return writeJSON(ctx, fiber.StatusOK, asset)
}

// getAssetURL returns a signed read URL.
func (handler handler) getAssetURL(ctx *fiber.Ctx) error {
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
		httpguard.Object(groupsdomain.PermissionAssetsView, groupsdomain.ObjectAsset, id),
	); err != nil {
		return err
	}
	url, err := handler.services.Assets.GetURL(ctx.UserContext(), id, 0)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, assetURLResponse{URL: url})
}

// listAssets lists assets.
func (handler handler) listAssets(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.All(groupsdomain.PermissionAssetsView, groupsdomain.ObjectAsset),
	); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := assetFilterFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Assets.List(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, assetListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
		Query:         filter.Query.String(),
		Sort:          filter.Sort.Key,
		Direction:     string(filter.Sort.Direction),
	})
}

// listNamespaces lists active asset namespaces.
func (handler handler) listNamespaces(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.All(groupsdomain.PermissionAssetsView, groupsdomain.ObjectAsset),
	); err != nil {
		return err
	}
	namespaces, err := handler.services.Assets.ListNamespaces(ctx.UserContext())
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, namespaceListResponse{Namespaces: namespaces})
}

// listFolders lists virtual folders.
func (handler handler) listFolders(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.All(groupsdomain.PermissionAssetsView, groupsdomain.ObjectAsset),
	); err != nil {
		return err
	}
	folders, err := handler.services.Assets.ListFolders(ctx.UserContext(), port.FolderFilter{
		Namespace:  domain.Namespace(ctx.Query("namespace")),
		PathPrefix: domain.VirtualPath(ctx.Query("path_prefix")),
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, folderListResponse{Folders: folders})
}
