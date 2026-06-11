package harness

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/realmkit/rk-backend/pkg/storage"
)

// MemoryStorage stores S3-compatible objects in memory for e2e tests.
type MemoryStorage struct {
	mu      sync.RWMutex
	objects map[string]memoryObject
	baseURL string
	now     func() time.Time
}

// MemoryStorageOption changes in-memory storage behavior.
type MemoryStorageOption func(*MemoryStorage)

// WithMemoryStorageBaseURL changes generated presigned URLs.
func WithMemoryStorageBaseURL(baseURL string) MemoryStorageOption {
	return func(store *MemoryStorage) {
		store.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// NewMemoryStorage creates an empty in-memory object store.
func NewMemoryStorage(options ...MemoryStorageOption) *MemoryStorage {
	store := &MemoryStorage{
		objects: map[string]memoryObject{},
		baseURL: "https://storage.e2e",
		now:     func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(store)
	}
	if store.baseURL == "" {
		store.baseURL = "https://storage.e2e"
	}
	return store
}

// Health verifies the in-memory storage backend is available.
func (store *MemoryStorage) Health(context.Context) error {
	return nil
}

// Put stores object bytes.
func (store *MemoryStorage) Put(
	_ context.Context,
	object storage.Object,
	body io.Reader,
) (storage.StoredObject, error) {
	if body == nil {
		body = strings.NewReader("")
	}
	payload, err := io.ReadAll(body)
	if err != nil {
		return storage.StoredObject{}, fmt.Errorf("read object body: %w", err)
	}
	return store.seed(object, payload), nil
}

// Delete deletes an object by key.
func (store *MemoryStorage) Delete(_ context.Context, key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.objects, key)
	return nil
}

// PresignPut creates a deterministic upload request.
func (store *MemoryStorage) PresignPut(
	_ context.Context,
	request storage.PresignPutRequest,
) (storage.PresignedRequest, error) {
	headers := map[string]string{}
	if request.ContentType != "" {
		headers[http.CanonicalHeaderKey("content-type")] = request.ContentType
	}
	return storage.PresignedRequest{
		Method:    http.MethodPut,
		URL:       store.url("upload", request.Key),
		Headers:   headers,
		ExpiresAt: store.now().Add(request.ExpiresIn),
	}, nil
}

// PresignGet creates a deterministic download URL.
func (store *MemoryStorage) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return store.url("read", key), nil
}

// Head returns object metadata.
func (store *MemoryStorage) Head(_ context.Context, key string) (storage.ObjectInfo, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	object, ok := store.objects[key]
	if !ok {
		return storage.ObjectInfo{}, fmt.Errorf("memory storage object %q not found", key)
	}
	return object.info(), nil
}

// Seed stores object bytes directly for setup-heavy e2e tests.
func (store *MemoryStorage) Seed(object storage.Object, payload []byte) storage.StoredObject {
	return store.seed(object, payload)
}

// Count returns the number of stored objects.
func (store *MemoryStorage) Count() int {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return len(store.objects)
}

// seed stores an object without reading from an io.Reader.
func (store *MemoryStorage) seed(object storage.Object, payload []byte) storage.StoredObject {
	stored := storage.StoredObject{Key: object.Key, ETag: memoryETag(payload)}
	info := storage.ObjectInfo{
		Key:         object.Key,
		ETag:        stored.ETag,
		ContentType: object.ContentType,
		SizeBytes:   int64(len(payload)),
		Metadata:    cloneMetadata(object.Metadata),
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	store.objects[object.Key] = memoryObject{infoValue: info, payload: append([]byte(nil), payload...)}
	return stored
}

// url returns a deterministic object URL.
func (store *MemoryStorage) url(kind string, key string) string {
	return store.baseURL + "/" + kind + "/" + url.PathEscape(key)
}

// memoryObject contains one stored test object.
type memoryObject struct {
	infoValue storage.ObjectInfo
	payload   []byte
}

// info returns a defensive copy of object metadata.
func (object memoryObject) info() storage.ObjectInfo {
	info := object.infoValue
	info.Metadata = cloneMetadata(info.Metadata)
	return info
}

// cloneMetadata copies provider metadata.
func cloneMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	clone := make(map[string]string, len(metadata))
	for key, value := range metadata {
		clone[key] = value
	}
	return clone
}

// memoryETag returns a stable test ETag for payload.
func memoryETag(payload []byte) string {
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}
