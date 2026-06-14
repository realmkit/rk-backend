package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"github.com/realmkit/rk-backend/pkg/search"
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

// TestUserRepositoryListSearchesClaimCache verifies user list search uses cached provider claims.
func TestUserRepositoryListSearchesClaimCache(t *testing.T) {
	users, _, claims := newRepositories(t)
	ian := createUserWithClaims(t, users, claims, "ian", "ian@example.test", "Ian Castano")
	createUserWithClaims(t, users, claims, "ada", "ada@example.test", "Ada Lovelace")

	result, err := users.List(context.Background(), userFilter(t, "castano", "email", "asc"), page(t, 10, ""))
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].User.ID != ian.ID {
		t.Fatalf("Items = %+v, want Ian result", result.Items)
	}
	if result.Items[0].Claims == nil || result.Items[0].Claims.Email != "ian@example.test" {
		t.Fatalf("Claims = %+v, want cached email", result.Items[0].Claims)
	}
}

// TestUserRepositoryFindSummariesByIDsReturnsClaimCache verifies batch summaries.
func TestUserRepositoryFindSummariesByIDsReturnsClaimCache(t *testing.T) {
	users, _, claims := newRepositories(t)
	ian := createUserWithClaims(t, users, claims, "ian", "ian@example.test", "Ian Castano")
	createUserWithClaims(t, users, claims, "ada", "ada@example.test", "Ada Lovelace")

	result, err := users.FindSummariesByIDs(context.Background(), []uuid.UUID{ian.ID, ian.ID})
	if err != nil {
		t.Fatalf("FindSummariesByIDs() error = %v", err)
	}
	if len(result) != 1 || result[ian.ID].User.ID != ian.ID {
		t.Fatalf("FindSummariesByIDs() = %+v, want only Ian", result)
	}
	if result[ian.ID].Claims == nil || result[ian.ID].Claims.DisplayName != "Ian Castano" {
		t.Fatalf("Claims = %+v, want cached display name", result[ian.ID].Claims)
	}
}

// TestUserRepositoryListUsesCursorPagination verifies user list cursors are stable.
func TestUserRepositoryListUsesCursorPagination(t *testing.T) {
	users, _, claims := newRepositories(t)
	first := createUserWithClaims(t, users, claims, "ada", "ada@example.test", "Ada Lovelace")
	second := createUserWithClaims(t, users, claims, "ian", "ian@example.test", "Ian Castano")

	filter := userFilter(t, "", "email", "asc")
	firstPage, err := users.List(context.Background(), filter, page(t, 1, ""))
	if err != nil {
		t.Fatalf("List() first page error = %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].User.ID != first.ID || firstPage.NextCursor == "" {
		t.Fatalf("firstPage = %+v, want Ada plus cursor", firstPage)
	}

	secondPage, err := users.List(context.Background(), filter, page(t, 1, firstPage.NextCursor))
	if err != nil {
		t.Fatalf("List() second page error = %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].User.ID != second.ID {
		t.Fatalf("secondPage = %+v, want Ian", secondPage)
	}
}

// TestUserRepositoryListRejectsInvalidCursor verifies malformed cursors fail.
func TestUserRepositoryListRejectsInvalidCursor(t *testing.T) {
	users, _, _ := newRepositories(t)
	_, err := users.List(context.Background(), userFilter(t, "", "email", "asc"), page(t, 1, "not-a-cursor"))
	if !errors.Is(err, search.ErrInvalidCursor) {
		t.Fatalf("List() error = %v, want invalid cursor", err)
	}
}

// TestIdentityLinkRepositoryLifecycle verifies identity link persistence.
func TestIdentityLinkRepositoryLifecycle(t *testing.T) {
	_, links, _ := newRepositories(t)
	now := time.Now().UTC()
	link := domain.IdentityLink{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Provider:    "generic_oidc",
		Issuer:      "issuer",
		Subject:     "subject",
		SubjectHash: "hash",
		LinkedAt:    now,
	}
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
	cache := domain.ClaimCache{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Issuer:   "issuer",
		Subject:  "subject",
		Username: "ian",
		SyncedAt: time.Now().UTC(),
	}
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

// createUserWithClaims stores a user and matching provider claims.
func createUserWithClaims(
	t *testing.T,
	users UserRepository,
	claims ClaimCacheRepository,
	username string,
	email string,
	displayName string,
) domain.User {
	t.Helper()
	user, err := users.Create(context.Background(), testUser())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = claims.Upsert(context.Background(), domain.ClaimCache{
		ID:            uuid.New(),
		UserID:        user.ID,
		Issuer:        "issuer",
		Subject:       user.ID.String(),
		Username:      username,
		Email:         email,
		EmailVerified: true,
		DisplayName:   displayName,
		ClaimsHash:    username + "-hash",
		SyncedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	return user
}

// userFilter returns a validated list filter.
func userFilter(t *testing.T, query string, sortKey string, direction string) port.UserFilter {
	t.Helper()
	text, err := search.NewTextQuery(query, search.QueryOptions{})
	if err != nil {
		t.Fatalf("NewTextQuery() error = %v", err)
	}
	sort, err := search.NewSort(sortKey, direction, port.DefaultUserSort(), port.AllowedUserSorts())
	if err != nil {
		t.Fatalf("NewSort() error = %v", err)
	}
	return port.UserFilter{Query: text, Sort: sort}
}

// page returns normalized pagination options.
func page(t *testing.T, limit int, cursor string) pagination.Page {
	t.Helper()
	page, err := pagination.New(pagination.Request{Limit: limit, Cursor: cursor})
	if err != nil {
		t.Fatalf("pagination.New() error = %v", err)
	}
	return page
}
