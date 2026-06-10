package application

import (
	"context"

	"github.com/niflaot/gamehub-go/module/assets/domain"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
)

const (
	// assetCreatedEvent is emitted when an asset row is created.
	assetCreatedEvent eventdomain.EventKey = "assets.asset.created"

	// assetUpdatedEvent is emitted when asset metadata changes.
	assetUpdatedEvent eventdomain.EventKey = "assets.asset.updated"

	// assetDeletedEvent is emitted when an asset is soft deleted.
	assetDeletedEvent eventdomain.EventKey = "assets.asset.deleted"

	// assetUploadCompletedEvent is emitted when upload completion succeeds.
	assetUploadCompletedEvent eventdomain.EventKey = "assets.asset.upload_completed"
)

// publishAssetEvent publishes one asset lifecycle event.
func (service Service) publishAssetEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	asset domain.Asset,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:            key,
		SchemaVersion:  1,
		Producer:       eventdomain.ProducerAssets,
		AggregateType:  "asset",
		AggregateID:    emitter.UUID(asset.ID),
		ActorUserID:    asset.CreatedByUserID,
		Payload:        assetPayload(asset),
		Scopes:         assetScopes(asset),
		IdempotencyKey: asset.ID.String() + ":" + string(key),
	})
}

// assetPayload returns a safe asset event payload.
func assetPayload(asset domain.Asset) map[string]any {
	return map[string]any{
		"id":           asset.ID,
		"namespace":    asset.Namespace,
		"path":         asset.Path,
		"filename":     asset.Filename,
		"visibility":   asset.Visibility,
		"status":       asset.Status,
		"content_type": asset.ContentType,
		"size_bytes":   asset.SizeBytes,
		"version":      asset.Version,
	}
}

// assetScopes returns audience scopes for one asset.
func assetScopes(asset domain.Asset) []eventdomain.Scope {
	scopes := []eventdomain.Scope{
		{Type: eventdomain.ScopeAsset, ID: asset.ID.String()},
	}
	if asset.CreatedByUserID != nil {
		scopes = append(scopes, eventdomain.Scope{
			Type: eventdomain.ScopeUser,
			ID:   asset.CreatedByUserID.String(),
		})
	}
	return scopes
}
