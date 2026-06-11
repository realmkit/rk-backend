// Package storage_e2e verifies storage infrastructure through e2e fixtures.
package storage_e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// TestStorageStoreContract verifies the e2e store satisfies storage behavior.
func TestStorageStoreContract(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start ecosystem with in-memory storage")
	ecosystem := harness.New(t)

	steps.Log("write object through storage.Store interface")
	var store storage.Store = ecosystem.Storage
	object := storage.Object{Key: "assets/logo.png", ContentType: "image/png"}
	stored, err := store.Put(context.Background(), object, strings.NewReader("logo"))
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if stored.ETag == "" {
		t.Fatalf("ETag = empty")
	}

	steps.Log("read object metadata")
	info, err := store.Head(context.Background(), object.Key)
	if err != nil {
		t.Fatalf("Head() error = %v", err)
	}
	if info.SizeBytes != 4 || info.ContentType != "image/png" {
		t.Fatalf("info = %+v, want object metadata", info)
	}

	steps.Log("create signed read URL")
	readURL, err := store.PresignGet(context.Background(), object.Key, 0)
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}
	if readURL == "" {
		t.Fatalf("read URL = empty")
	}
}
