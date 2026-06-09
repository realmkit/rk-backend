package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/user/domain"
	"github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUserRepositoryLifecycle verifies local user persistence.
func TestUserRepositoryLifecycle(t *testing.T) {
	users, _, _ := newRepositories(t)
	user, err := users.Create(context.Background(), testUser())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	avatarID := uuid.New()
	user.AvatarAssetID = &avatarID
	updated, err := users.Update(context.Background(), user, user.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 || updated.AvatarAssetID == nil || *updated.AvatarAssetID != avatarID {
		t.Fatalf("updated = %+v, want avatar and version", updated)
	}
	if err := users.TouchLastSeen(context.Background(), updated.ID); err != nil {
		t.Fatalf("TouchLastSeen() error = %v", err)
	}
	if _, err := users.FindByID(context.Background(), uuid.New()); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want not found", err)
	}
}

// TestIdentityLinkRepositoryLifecycle verifies identity link persistence.
func TestIdentityLinkRepositoryLifecycle(t *testing.T) {
	_, links, _ := newRepositories(t)
	now := time.Now().UTC()
	link := domain.IdentityLink{ID: uuid.New(), UserID: uuid.New(), Provider: "generic_oidc", Issuer: "issuer", Subject: "subject", SubjectHash: "hash", LinkedAt: now}
	created, err := links.Create(context.Background(), link)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := links.Create(context.Background(), link); !errors.Is(err, port.ErrConflict) {
		t.Fatalf("Create() duplicate error = %v, want conflict", err)
	}
	found, err := links.FindByIssuerSubject(context.Background(), created.Issuer, created.Subject)
	if err != nil {
		t.Fatalf("FindByIssuerSubject() error = %v", err)
	}
	seen := now.Add(time.Minute)
	found.LastSeenAt = &seen
	if err := links.Touch(context.Background(), found); err != nil {
		t.Fatalf("Touch() error = %v", err)
	}
}

// TestClaimCacheRepositoryUpsert verifies claim cache upserts.
func TestClaimCacheRepositoryUpsert(t *testing.T) {
	_, _, claims := newRepositories(t)
	cache := domain.ClaimCache{ID: uuid.New(), UserID: uuid.New(), Issuer: "issuer", Subject: "subject", Username: "ian", SyncedAt: time.Now().UTC()}
	created, err := claims.Upsert(context.Background(), cache)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	created.Username = "ian2"
	updated, err := claims.Upsert(context.Background(), created)
	if err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}
	if updated.Username != "ian2" {
		t.Fatalf("Username = %q, want ian2", updated.Username)
	}
	found, err := claims.FindByUserID(context.Background(), updated.UserID)
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if found.ID != updated.ID {
		t.Fatalf("FindByUserID() ID = %s, want %s", found.ID, updated.ID)
	}
}

// newRepositories creates migrated user repositories.
func newRepositories(t *testing.T) (UserRepository, IdentityLinkRepository, ClaimCacheRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	store := orm.NewStore(db)
	return NewUserRepository(store), NewIdentityLinkRepository(store), NewClaimCacheRepository(store)
}

// testUser returns a valid local user.
func testUser() domain.User {
	return domain.User{ID: uuid.New(), Status: domain.StatusActive, FirstSeenAt: time.Now().UTC(), Version: 1}
}
