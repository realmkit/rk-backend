package postgres

import (
	"github.com/niflaot/gamehub-go/module/assets/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// assetModelFromDomain maps domain asset to persistence.
func assetModelFromDomain(asset domain.Asset) AssetModel {
	return AssetModel{
		ID:              orm.ID{ID: asset.ID},
		Namespace:       string(asset.Namespace),
		Path:            string(asset.Path),
		Filename:        string(asset.Filename),
		DisplayName:     asset.DisplayName,
		Visibility:      string(asset.Visibility),
		Status:          string(asset.Status),
		StorageKey:      asset.StorageKey,
		Bucket:          asset.Bucket,
		ContentType:     asset.ContentType,
		SizeBytes:       asset.SizeBytes,
		ETag:            asset.ETag,
		CreatedByUserID: asset.CreatedByUserID,
		Version:         asset.Version,
	}
}

// assetFromModel maps persistence asset to domain.
func assetFromModel(model AssetModel) domain.Asset {
	return domain.Asset{
		ID:              model.ID.ID,
		Namespace:       domain.Namespace(model.Namespace),
		Path:            domain.VirtualPath(model.Path),
		Filename:        domain.Filename(model.Filename),
		DisplayName:     model.DisplayName,
		Visibility:      domain.Visibility(model.Visibility),
		Status:          domain.Status(model.Status),
		StorageKey:      model.StorageKey,
		Bucket:          model.Bucket,
		ContentType:     model.ContentType,
		SizeBytes:       model.SizeBytes,
		ETag:            model.ETag,
		CreatedByUserID: model.CreatedByUserID,
		Version:         model.Version,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
