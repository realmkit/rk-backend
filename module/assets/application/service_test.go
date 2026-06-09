package application

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/assets/domain"
	"github.com/niflaot/gamehub-go/module/assets/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/storage"
)

// TestServiceCreateUploadIntentCreatesAssetAndSignedURL verifies intent creation.
func TestServiceCreateUploadIntentCreatesAssetAndSignedURL(t *testing.T) {
	repository := newMemoryRepository()
	store := &memoryStore{}
	service := NewService(repository, store, "gamehub-assets")

	intent, err := service.CreateUploadIntent(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("CreateUploadIntent() error = %v", err)
	}
	if intent.Asset.Status != domain.StatusPendingUpload || intent.URL == "" || intent.Method != "PUT" {
		t.Fatalf("intent = %+v, want pending asset and upload request", intent)
	}
	if repository.created == 0 || store.presignedPut == 0 {
		t.Fatalf("created=%d presigned=%d, want both called", repository.created, store.presignedPut)
	}
}

// TestServiceCompleteUploadConfirmsStorageObject verifies pending uploads become available.
func TestServiceCompleteUploadConfirmsStorageObject(t *testing.T) {
	repository := newMemoryRepository()
	store := &memoryStore{}
	service := NewService(repository, store, "gamehub-assets")
	intent, err := service.CreateUploadIntent(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("CreateUploadIntent() error = %v", err)
	}
	store.info = storage.ObjectInfo{Key: intent.Asset.StorageKey, ETag: "etag", ContentType: "image/png", SizeBytes: 512}

	asset, err := service.CompleteUpload(context.Background(), port.CompleteUploadCommand{ID: intent.Asset.ID})
	if err != nil {
		t.Fatalf("CompleteUpload() error = %v", err)
	}
	if asset.Status != domain.StatusAvailable || asset.ETag != "etag" || asset.Version != 2 {
		t.Fatalf("asset = %+v, want available version 2", asset)
	}
}

// TestServiceCompleteUploadIsIdempotentForAvailableAsset verifies available assets return without storage checks.
func TestServiceCompleteUploadIsIdempotentForAvailableAsset(t *testing.T) {
	repository := newMemoryRepository()
	store := &memoryStore{}
	service := NewService(repository, store, "gamehub-assets")
	intent, err := service.CreateUploadIntent(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("CreateUploadIntent() error = %v", err)
	}
	asset := intent.Asset
	asset.Status = domain.StatusAvailable
	repository.items[asset.ID] = asset

	found, err := service.CompleteUpload(context.Background(), port.CompleteUploadCommand{ID: asset.ID})
	if err != nil {
		t.Fatalf("CompleteUpload() error = %v", err)
	}
	if found.ID != asset.ID || store.heads != 0 {
		t.Fatalf("found=%s heads=%d, want idempotent return without head", found.ID, store.heads)
	}
}

// TestServiceCompleteUploadRejectsMismatchedObject verifies storage metadata is enforced.
func TestServiceCompleteUploadRejectsMismatchedObject(t *testing.T) {
	repository := newMemoryRepository()
	store := &memoryStore{info: storage.ObjectInfo{ContentType: "image/png", SizeBytes: 1}}
	service := NewService(repository, store, "gamehub-assets")
	intent, err := service.CreateUploadIntent(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("CreateUploadIntent() error = %v", err)
	}

	_, err = service.CompleteUpload(context.Background(), port.CompleteUploadCommand{ID: intent.Asset.ID})
	if !errors.Is(err, port.ErrUploadMismatch) {
		t.Fatalf("CompleteUpload() error = %v, want %v", err, port.ErrUploadMismatch)
	}
}

// TestServiceGetURLRequiresAvailableAsset verifies signed reads only work for available assets.
func TestServiceGetURLRequiresAvailableAsset(t *testing.T) {
	repository := newMemoryRepository()
	store := &memoryStore{}
	service := NewService(repository, store, "gamehub-assets")
	intent, err := service.CreateUploadIntent(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("CreateUploadIntent() error = %v", err)
	}

	_, err = service.GetURL(context.Background(), intent.Asset.ID, 0)
	if !errors.Is(err, port.ErrInvalidState) {
		t.Fatalf("GetURL() error = %v, want %v", err, port.ErrInvalidState)
	}
}

// TestServiceMutableOperationsUseRepository verifies read, list, folder, update, and delete paths.
func TestServiceMutableOperationsUseRepository(t *testing.T) {
	repository := newMemoryRepository()
	store := &memoryStore{}
	service := NewService(repository, store, "gamehub-assets")
	intent, err := service.CreateUploadIntent(context.Background(), validCommand())
	if err != nil {
		t.Fatalf("CreateUploadIntent() error = %v", err)
	}
	got, err := service.Get(context.Background(), intent.Asset.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != intent.Asset.ID {
		t.Fatalf("Get() ID = %s, want %s", got.ID, intent.Asset.ID)
	}
	list, err := service.List(context.Background(), port.AssetFilter{PathPrefix: " /brand/ "}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	folders, err := service.ListFolders(context.Background(), port.FolderFilter{PathPrefix: " /brand/ "})
	if err != nil {
		t.Fatalf("ListFolders() error = %v", err)
	}
	if len(folders) != 1 {
		t.Fatalf("ListFolders() = %v, want one folder", folders)
	}
	updated, err := service.Update(context.Background(), port.UpdateAssetCommand{ID: intent.Asset.ID, DisplayName: "Updated", Path: "brand/updated", Visibility: domain.VisibilityAuthenticated, ExpectedVersion: intent.Asset.Version})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.DisplayName != "Updated" || updated.Version != 2 {
		t.Fatalf("updated = %+v, want updated version 2", updated)
	}
	if err := service.Delete(context.Background(), port.DeleteAssetCommand{ID: updated.ID, ExpectedVersion: updated.Version}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

// validCommand returns a valid upload intent command.
func validCommand() port.CreateUploadIntentCommand {
	return port.CreateUploadIntentCommand{
		Namespace:   "community",
		Path:        "brand/icons",
		Filename:    "logo.png",
		Visibility:  domain.VisibilityPublic,
		ContentType: "image/png",
		SizeBytes:   512,
	}
}

// newMemoryRepository creates an in-memory asset repository.
func newMemoryRepository() *memoryRepository {
	return &memoryRepository{items: map[uuid.UUID]domain.Asset{}}
}

// memoryRepository stores assets in memory for tests.
type memoryRepository struct {
	items   map[uuid.UUID]domain.Asset
	created int
}

// Create stores an asset.
func (repository *memoryRepository) Create(_ context.Context, asset domain.Asset) (domain.Asset, error) {
	repository.created++
	repository.items[asset.ID] = asset
	return asset, nil
}

// Update stores mutable asset fields.
func (repository *memoryRepository) Update(_ context.Context, asset domain.Asset, expectedVersion uint64) (domain.Asset, error) {
	current := repository.items[asset.ID]
	if current.Version != expectedVersion {
		return domain.Asset{}, port.ErrPreconditionFailed
	}
	asset.Version = expectedVersion + 1
	repository.items[asset.ID] = asset
	return asset, nil
}

// FindByID returns one asset.
func (repository *memoryRepository) FindByID(_ context.Context, id uuid.UUID) (domain.Asset, error) {
	asset, ok := repository.items[id]
	if !ok {
		return domain.Asset{}, port.ErrNotFound
	}
	return asset, nil
}

// List returns matching assets.
func (repository *memoryRepository) List(context.Context, port.AssetFilter, pagination.Page) (pagination.Result[domain.Asset], error) {
	items := make([]domain.Asset, 0, len(repository.items))
	for _, asset := range repository.items {
		items = append(items, asset)
	}
	return pagination.Result[domain.Asset]{Items: items}, nil
}

// ListFolders returns direct child folders.
func (repository *memoryRepository) ListFolders(context.Context, port.FolderFilter) ([]string, error) {
	return []string{"brand"}, nil
}

// Delete soft deletes one asset.
func (repository *memoryRepository) Delete(_ context.Context, id uuid.UUID, expectedVersion uint64) error {
	current := repository.items[id]
	if current.Version != expectedVersion {
		return port.ErrPreconditionFailed
	}
	delete(repository.items, id)
	return nil
}

// memoryStore stores objects in memory for tests.
type memoryStore struct {
	info         storage.ObjectInfo
	heads        int
	presignedPut int
}

// Health verifies the storage backend is reachable.
func (store *memoryStore) Health(context.Context) error {
	return nil
}

// Put stores object bytes.
func (store *memoryStore) Put(context.Context, storage.Object, io.Reader) (storage.StoredObject, error) {
	return storage.StoredObject{}, nil
}

// Delete deletes an object by key.
func (store *memoryStore) Delete(context.Context, string) error {
	return nil
}

// PresignPut creates a presigned upload request.
func (store *memoryStore) PresignPut(_ context.Context, request storage.PresignPutRequest) (storage.PresignedRequest, error) {
	store.presignedPut++
	return storage.PresignedRequest{Method: "PUT", URL: "https://storage.test/" + request.Key, Headers: map[string]string{"Content-Type": request.ContentType}, ExpiresAt: time.Now().Add(time.Minute)}, nil
}

// PresignGet creates a presigned download URL.
func (store *memoryStore) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "https://storage.test/read", nil
}

// Head returns object metadata.
func (store *memoryStore) Head(context.Context, string) (storage.ObjectInfo, error) {
	store.heads++
	return store.info, nil
}
