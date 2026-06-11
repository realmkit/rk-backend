package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/assets/domain"
	"github.com/niflaot/gamehub-go/module/assets/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAssetRepositoryLifecycle verifies create, read, list, update, folders, and delete.
func TestAssetRepositoryLifecycle(t *testing.T) {
	repository := newAssetRepository(t)
	asset, err := repository.Create(context.Background(), testAsset())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	found, err := repository.FindByID(context.Background(), asset.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.StorageKey != asset.StorageKey {
		t.Fatalf("FindByID() StorageKey = %q, want %q", found.StorageKey, asset.StorageKey)
	}
	asset.DisplayName = "Updated Logo"
	asset.Visibility = domain.VisibilityAuthenticated
	updated, err := repository.Update(context.Background(), asset, asset.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 || updated.DisplayName != "Updated Logo" {
		t.Fatalf("updated = %+v, want version 2 and display name", updated)
	}
	list, err := repository.List(
		context.Background(),
		port.AssetFilter{Namespace: "community", PathPrefix: "brand"},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	folders, err := repository.ListFolders(context.Background(), port.FolderFilter{Namespace: "community"})
	if err != nil {
		t.Fatalf("ListFolders() error = %v", err)
	}
	if len(folders) != 1 || folders[0] != "brand" {
		t.Fatalf("ListFolders() = %v, want brand", folders)
	}
	if err := repository.Delete(context.Background(), updated.ID, updated.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.FindByID(context.Background(), updated.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() after delete error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestAssetRepositoryRejectsStaleVersion verifies optimistic concurrency.
func TestAssetRepositoryRejectsStaleVersion(t *testing.T) {
	repository := newAssetRepository(t)
	asset, err := repository.Create(context.Background(), testAsset())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	asset.DisplayName = "Updated"

	_, err = repository.Update(context.Background(), asset, 999)
	if !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("Update() error = %v, want %v", err, port.ErrPreconditionFailed)
	}
}

// TestAssetRepositoryListIsBounded verifies hot asset lists do not return unbounded rows.
func TestAssetRepositoryListIsBounded(t *testing.T) {
	repository := newAssetRepository(t)
	for index := 0; index < 3; index++ {
		asset := testAsset()
		asset.ID = uuid.New()
		asset.Filename = domain.Filename(uuid.NewString() + ".png")
		asset.StorageKey = "assets/community/" + uuid.NewString() + "/logo.png"
		if _, err := repository.Create(context.Background(), asset); err != nil {
			t.Fatalf("Create(%d) error = %v", index, err)
		}
	}

	list, err := repository.List(
		context.Background(),
		port.AssetFilter{Namespace: "community"},
		pagination.Page{Limit: 1},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 || list.NextCursor == "" {
		t.Fatalf("List() = %+v, want one bounded item and next cursor", list)
	}
}

// newAssetRepository creates a migrated repository.
func newAssetRepository(t *testing.T) AssetRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	return NewAssetRepository(orm.NewStore(db))
}

// testAsset returns a valid asset.
func testAsset() domain.Asset {
	return domain.Asset{
		ID:          uuid.New(),
		Namespace:   "community",
		Path:        "brand/icons",
		Filename:    "logo.png",
		DisplayName: "Logo",
		Visibility:  domain.VisibilityPublic,
		Status:      domain.StatusPendingUpload,
		StorageKey:  "assets/community/2026/06/" + uuid.NewString() + "/logo.png",
		Bucket:      "gamehub-assets",
		ContentType: "image/png",
		SizeBytes:   512,
		Version:     1,
	}
}
