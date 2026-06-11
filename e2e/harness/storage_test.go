package harness

import (
	"context"
	"strings"
	"testing"

	"github.com/niflaot/gamehub-go/pkg/storage"
)

// TestMemoryStoragePutHeadAndDelete verifies in-memory object lifecycle behavior.
func TestMemoryStoragePutHeadAndDelete(t *testing.T) {
	store := NewMemoryStorage()
	object := storage.Object{
		Key:         "community/banner.png",
		ContentType: "image/png",
		Metadata:    map[string]string{"owner": "e2e"},
	}

	stored, err := store.Put(context.Background(), object, strings.NewReader("image"))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if stored.Key != object.Key || stored.ETag == "" {
		t.Fatalf("stored = %+v, want key and etag", stored)
	}

	info, err := store.Head(context.Background(), object.Key)
	if err != nil {
		t.Fatalf("Head() error = %v", err)
	}
	if info.ContentType != object.ContentType || info.SizeBytes != 5 || info.Metadata["owner"] != "e2e" {
		t.Fatalf("info = %+v, want stored metadata", info)
	}

	if err := store.Delete(context.Background(), object.Key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Head(context.Background(), object.Key); err == nil {
		t.Fatalf("Head() error = nil, want missing object error")
	}
}

// TestMemoryStoragePresignsDeterministicURLs verifies signed URL shape.
func TestMemoryStoragePresignsDeterministicURLs(t *testing.T) {
	store := NewMemoryStorage(WithMemoryStorageBaseURL("https://cdn.e2e/"))

	put, err := store.PresignPut(
		context.Background(),
		storage.PresignPutRequest{Key: "avatars/user.png", ContentType: "image/png"},
	)
	if err != nil {
		t.Fatalf("PresignPut() error = %v", err)
	}
	if put.Method != "PUT" || put.URL != "https://cdn.e2e/upload/avatars%2Fuser.png" {
		t.Fatalf("put = %+v, want deterministic upload URL", put)
	}
	if put.Headers["Content-Type"] != "image/png" {
		t.Fatalf("Content-Type header = %q, want image/png", put.Headers["Content-Type"])
	}

	getURL, err := store.PresignGet(context.Background(), "avatars/user.png", 0)
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}
	if getURL != "https://cdn.e2e/read/avatars%2Fuser.png" {
		t.Fatalf("PresignGet() = %q, want deterministic read URL", getURL)
	}
}
