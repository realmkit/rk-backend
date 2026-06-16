package application

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// uploadIntentTTL is the default signed upload lifetime.
const uploadIntentTTL = 15 * time.Minute

// readURLTTL is the default signed read lifetime.
const readURLTTL = 15 * time.Minute

// Service manages assets.
type Service struct {
	repository port.AssetRepository
	store      storage.Store
	bucket     string
	clock      func() time.Time
	events     emitter.Publisher
}

// NewService creates an assets service.
func NewService(repository port.AssetRepository, store storage.Store, bucket string) Service {
	return Service{
		repository: repository,
		store:      store,
		bucket:     bucket,
		clock:      func() time.Time { return time.Now().UTC() },
	}
}

// WithEvents returns a copy of service that publishes asset events.
func (service Service) WithEvents(events emitter.Publisher) Service {
	service.events = events
	return service
}

// CreateUploadIntent creates an asset and presigned upload URL.
func (service Service) CreateUploadIntent(ctx context.Context, command port.CreateUploadIntentCommand) (port.UploadIntent, error) {
	asset := service.assetFromCommand(command)
	if err := asset.Validate(); err != nil {
		return port.UploadIntent{}, err
	}
	created, err := service.repository.Create(ctx, asset)
	if err != nil {
		return port.UploadIntent{}, err
	}
	signed, err := service.store.PresignPut(
		ctx,
		storage.PresignPutRequest{Key: created.StorageKey, ContentType: created.ContentType, ExpiresIn: uploadIntentTTL},
	)
	if err != nil {
		return port.UploadIntent{}, err
	}
	if err := service.publishAssetEvent(ctx, assetCreatedEvent, created); err != nil {
		return port.UploadIntent{}, err
	}
	return port.UploadIntent{
		Asset:     created,
		Method:    signed.Method,
		URL:       signed.URL,
		Headers:   signed.Headers,
		ExpiresAt: signed.ExpiresAt,
	}, nil
}

// CompleteUpload confirms the upload object exists.
func (service Service) CompleteUpload(ctx context.Context, command port.CompleteUploadCommand) (domain.Asset, error) {
	asset, err := service.repository.FindByID(ctx, command.ID)
	if err != nil {
		return domain.Asset{}, err
	}
	if asset.Status == domain.StatusAvailable {
		return asset, nil
	}
	if asset.Status != domain.StatusPendingUpload {
		return domain.Asset{}, port.ErrInvalidState
	}
	info, err := service.store.Head(ctx, asset.StorageKey)
	if err != nil {
		return domain.Asset{}, err
	}
	if !objectMatches(asset, info) {
		return domain.Asset{}, port.ErrUploadMismatch
	}
	asset.Status = domain.StatusAvailable
	asset.ETag = info.ETag
	updated, err := service.repository.Update(ctx, asset, asset.Version)
	if err != nil {
		return domain.Asset{}, err
	}
	return updated, service.publishAssetEvent(ctx, assetUploadCompletedEvent, updated)
}

// Get returns one asset.
func (service Service) Get(ctx context.Context, id uuid.UUID) (domain.Asset, error) {
	return service.repository.FindByID(ctx, id)
}

// GetURL returns a signed read URL.
func (service Service) GetURL(ctx context.Context, id uuid.UUID, ttl time.Duration) (string, error) {
	asset, err := service.repository.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if asset.Status != domain.StatusAvailable {
		return "", port.ErrInvalidState
	}
	if ttl <= 0 {
		ttl = readURLTTL
	}
	return service.store.PresignGet(ctx, asset.StorageKey, ttl)
}

// List returns matching assets.
func (service Service) List(ctx context.Context, filter port.AssetFilter, page pagination.Page) (pagination.Result[domain.Asset], error) {
	return service.repository.List(ctx, normalizeFilter(filter), page)
}

// ListNamespaces returns active asset namespaces.
func (service Service) ListNamespaces(ctx context.Context) ([]string, error) {
	return service.repository.ListNamespaces(ctx)
}

// ListFolders returns direct virtual folder children.
func (service Service) ListFolders(ctx context.Context, filter port.FolderFilter) ([]string, error) {
	filter.PathPrefix = domain.NormalizePath(filter.PathPrefix)
	return service.repository.ListFolders(ctx, filter)
}

// Update changes mutable asset fields.
func (service Service) Update(ctx context.Context, command port.UpdateAssetCommand) (domain.Asset, error) {
	asset, err := service.repository.FindByID(ctx, command.ID)
	if err != nil {
		return domain.Asset{}, err
	}
	asset.Namespace = command.Namespace
	asset.DisplayName = strings.TrimSpace(command.DisplayName)
	asset.Path = domain.NormalizePath(command.Path)
	asset.Visibility = command.Visibility
	asset = asset.Normalize()
	if err := asset.Validate(); err != nil {
		return domain.Asset{}, err
	}
	updated, err := service.repository.Update(ctx, asset, command.ExpectedVersion)
	if err != nil {
		return domain.Asset{}, err
	}
	return updated, service.publishAssetEvent(ctx, assetUpdatedEvent, updated)
}

// Delete soft deletes one asset.
func (service Service) Delete(ctx context.Context, command port.DeleteAssetCommand) error {
	asset, err := service.repository.FindByID(ctx, command.ID)
	if err != nil {
		return err
	}
	if err := service.repository.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.publishAssetEvent(ctx, assetDeletedEvent, asset)
}

// assetFromCommand creates an asset from command.
func (service Service) assetFromCommand(command port.CreateUploadIntentCommand) domain.Asset {
	id := uuid.New()
	asset := domain.Asset{
		ID:              id,
		Namespace:       command.Namespace,
		Path:            domain.NormalizePath(command.Path),
		Filename:        domain.NormalizeFilename(command.Filename),
		DisplayName:     strings.TrimSpace(command.DisplayName),
		Visibility:      command.Visibility,
		Status:          domain.StatusPendingUpload,
		Bucket:          service.bucket,
		ContentType:     strings.ToLower(strings.TrimSpace(command.ContentType)),
		SizeBytes:       command.SizeBytes,
		CreatedByUserID: command.CreatedByUserID,
		Version:         1,
	}
	asset.StorageKey = service.storageKey(asset, id)
	return asset.Normalize()
}

// storageKey returns the physical storage key.
func (service Service) storageKey(asset domain.Asset, id uuid.UUID) string {
	now := service.clock()
	filename := strings.ReplaceAll(string(asset.Filename), " ", "_")
	return path.Join("assets", string(asset.Namespace), now.Format("2006"), now.Format("01"), id.String(), filename)
}

// objectMatches reports whether stored object metadata matches intent.
func objectMatches(asset domain.Asset, info storage.ObjectInfo) bool {
	if info.SizeBytes != asset.SizeBytes {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(info.ContentType), strings.TrimSpace(asset.ContentType))
}

// normalizeFilter normalizes list filters.
func normalizeFilter(filter port.AssetFilter) port.AssetFilter {
	filter.Path = domain.NormalizePath(filter.Path)
	filter.PathPrefix = domain.NormalizePath(filter.PathPrefix)
	return filter
}

// Ensure Service implements port.Service.
var _ port.Service = Service{}
