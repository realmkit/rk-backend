package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/identity"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestServiceProvisionCreatesUserLinkAndClaims verifies lazy provisioning.
func TestServiceProvisionCreatesUserLinkAndClaims(t *testing.T) {
	service, users, links, claims := newTestService()
	external := testIdentity()
	principal, err := service.Provision(
		context.Background(),
		external,
		auth.Token{Identity: external, Audience: []string{"api"}, Scopes: []string{"openid"}},
	)
	if err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	if principal.UserID == uuid.Nil || len(users.items) != 1 || len(links.items) != 1 || len(claims.items) != 1 {
		t.Fatalf(
			"principal=%+v users=%d links=%d claims=%d, want provisioned",
			principal,
			len(users.items),
			len(links.items),
			len(claims.items),
		)
	}
}

// TestServicePublishesUserEvents verifies provisioning and updates emit events.
func TestServicePublishesUserEvents(t *testing.T) {
	events := &eventtesting.PublisherRecorder{}
	users := &memoryUsers{items: map[uuid.UUID]domain.User{}}
	links := &memoryLinks{items: map[string]domain.IdentityLink{}}
	claims := &memoryClaims{items: map[uuid.UUID]domain.ClaimCache{}}
	service := NewService(Dependencies{
		Users:        users,
		Links:        links,
		Claims:       claims,
		Transactions: fakeTx{},
		Provider:     "generic_oidc",
		Events:       events,
	})
	external := testIdentity()
	principal, err := service.Provision(context.Background(), external, auth.Token{Identity: external})
	if err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	if _, err := service.UpdateCurrent(context.Background(), port.UpdateCurrentCommand{
		UserID:          principal.UserID,
		ExpectedVersion: 1,
	}); err != nil {
		t.Fatalf("UpdateCurrent() error = %v", err)
	}
	assertUserEventKeys(t, events.Drafts(), []string{
		"users.user.provisioned",
		"users.identity.linked",
		"users.identity.claim_refreshed",
		"users.user.updated",
	})
}

// assertUserEventKeys verifies event draft key order.
func assertUserEventKeys(t *testing.T, drafts []eventdomain.Draft, want []string) {
	t.Helper()
	if len(drafts) != len(want) {
		t.Fatalf("event count = %d, want %d", len(drafts), len(want))
	}
	for index, key := range want {
		if string(drafts[index].Key) != key {
			t.Fatalf("event[%d] = %s, want %s", index, drafts[index].Key, key)
		}
	}
}

// TestServiceProvisionReusesExistingLink verifies repeated provisioning is idempotent.
func TestServiceProvisionReusesExistingLink(t *testing.T) {
	service, _, _, _ := newTestService()
	external := testIdentity()
	first, err := service.Provision(context.Background(), external, auth.Token{Identity: external})
	if err != nil {
		t.Fatalf("first Provision() error = %v", err)
	}
	second, err := service.Provision(context.Background(), external, auth.Token{Identity: external})
	if err != nil {
		t.Fatalf("second Provision() error = %v", err)
	}
	if second.UserID != first.UserID {
		t.Fatalf("second UserID = %s, want %s", second.UserID, first.UserID)
	}
}

// TestServiceProvisionRejectsDisabledUser verifies disabled local users cannot authenticate.
func TestServiceProvisionRejectsDisabledUser(t *testing.T) {
	service, users, _, _ := newTestService()
	external := testIdentity()
	principal, err := service.Provision(context.Background(), external, auth.Token{Identity: external})
	if err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	user := users.items[principal.UserID]
	user.Status = domain.StatusDisabled
	users.items[user.ID] = user
	_, err = service.Provision(context.Background(), external, auth.Token{Identity: external})
	if !errors.Is(err, auth.ErrDisabledUser) {
		t.Fatalf("Provision() error = %v, want disabled auth error", err)
	}
}

// TestServiceCurrentAndUpdate verifies current user aggregate and local updates.
func TestServiceCurrentAndUpdate(t *testing.T) {
	service, users, _, claims := newTestService()
	userID := uuid.New()
	users.items[userID] = domain.User{ID: userID, Status: domain.StatusActive, FirstSeenAt: time.Now().UTC(), Version: 1}
	claims.items[userID] = domain.ClaimCache{ID: uuid.New(), UserID: userID, Username: "ian", SyncedAt: time.Now().UTC()}
	avatarID := uuid.New()
	current, err := service.Current(context.Background(), userID)
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}
	if current.Claims == nil || current.Claims.Username != "ian" {
		t.Fatalf("Current() claims = %+v, want cache", current.Claims)
	}
	updated, err := service.UpdateCurrent(
		context.Background(),
		port.UpdateCurrentCommand{UserID: userID, AvatarAssetID: &avatarID, ExpectedVersion: 1},
	)
	if err != nil {
		t.Fatalf("UpdateCurrent() error = %v", err)
	}
	if updated.AvatarAssetID == nil || *updated.AvatarAssetID != avatarID || updated.Version != 2 {
		t.Fatalf("updated = %+v, want avatar and version", updated)
	}
}

// TestServiceGetReturnsUser verifies direct user lookup.
func TestServiceGetReturnsUser(t *testing.T) {
	service, users, _, _ := newTestService()
	userID := uuid.New()
	users.items[userID] = domain.User{ID: userID, Status: domain.StatusActive, FirstSeenAt: time.Now().UTC(), Version: 1}
	user, err := service.Get(context.Background(), userID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if user.ID != userID {
		t.Fatalf("Get() ID = %s, want %s", user.ID, userID)
	}
}

// TestServiceListReturnsUsers verifies list delegation.
func TestServiceListReturnsUsers(t *testing.T) {
	service, users, _, _ := newTestService()
	userID := uuid.New()
	users.items[userID] = domain.User{
		ID:          userID,
		Status:      domain.StatusActive,
		FirstSeenAt: time.Now().UTC(),
		Version:     1,
	}
	result, err := service.List(context.Background(), port.UserFilter{}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].User.ID != userID {
		t.Fatalf("List() = %+v, want seeded user", result.Items)
	}
}

// TestDevelopmentPrincipalRequiresExistingEnabledUser verifies development bypass guardrails.
func TestDevelopmentPrincipalRequiresExistingEnabledUser(t *testing.T) {
	service, users, _, _ := newTestService()
	userID := uuid.New()
	users.items[userID] = domain.User{ID: userID, Status: domain.StatusActive, FirstSeenAt: time.Now().UTC(), Version: 1}
	principal, err := service.DevelopmentPrincipal(context.Background(), userID)
	if err != nil {
		t.Fatalf("DevelopmentPrincipal() error = %v", err)
	}
	if !principal.DevelopmentBypass || principal.UserID != userID {
		t.Fatalf("principal = %+v, want development principal", principal)
	}
}

// TestDevelopmentPrincipalRejectsDisabledUser verifies disabled local user rejection.
func TestDevelopmentPrincipalRejectsDisabledUser(t *testing.T) {
	service, users, _, _ := newTestService()
	userID := uuid.New()
	users.items[userID] = domain.User{ID: userID, Status: domain.StatusDisabled, FirstSeenAt: time.Now().UTC(), Version: 1}
	_, err := service.DevelopmentPrincipal(context.Background(), userID)
	if !errors.Is(err, auth.ErrDisabledUser) {
		t.Fatalf("DevelopmentPrincipal() error = %v, want disabled auth error", err)
	}
}

// newTestService returns a service with memory repositories.
func newTestService() (Service, *memoryUsers, *memoryLinks, *memoryClaims) {
	users := &memoryUsers{items: map[uuid.UUID]domain.User{}}
	links := &memoryLinks{items: map[string]domain.IdentityLink{}}
	claims := &memoryClaims{items: map[uuid.UUID]domain.ClaimCache{}}
	service := NewService(Dependencies{Users: users, Links: links, Claims: claims, Transactions: fakeTx{}, Provider: "generic_oidc"})
	return service, users, links, claims
}

// testIdentity returns a provider identity.
func testIdentity() identity.ExternalIdentity {
	return identity.ExternalIdentity{Issuer: "https://auth.example", Subject: "subject", Username: "ian", RawClaimsHash: "hash"}
}

// fakeTx runs work without a real transaction.
type fakeTx struct{}

// WithinTx runs fn.
func (fakeTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// memoryUsers stores users in memory.
type memoryUsers struct {
	items map[uuid.UUID]domain.User
}

// Create stores a user.
func (repository *memoryUsers) Create(_ context.Context, user domain.User) (domain.User, error) {
	repository.items[user.ID] = user
	return user, nil
}

// Update stores mutable user fields.
func (repository *memoryUsers) Update(_ context.Context, user domain.User, expectedVersion uint64) (domain.User, error) {
	current := repository.items[user.ID]
	if current.Version != expectedVersion {
		return domain.User{}, port.ErrPreconditionFailed
	}
	user.Version = expectedVersion + 1
	repository.items[user.ID] = user
	return user, nil
}

// FindByID returns one user.
func (repository *memoryUsers) FindByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	user, ok := repository.items[id]
	if !ok {
		return domain.User{}, port.ErrNotFound
	}
	return user, nil
}

// List returns a memory user page.
func (repository *memoryUsers) List(
	_ context.Context,
	_ port.UserFilter,
	_ pagination.Page,
) (pagination.Result[port.UserSummary], error) {
	items := make([]port.UserSummary, 0, len(repository.items))
	for _, user := range repository.items {
		items = append(items, port.UserSummary{User: user})
	}
	return pagination.Result[port.UserSummary]{Items: items}, nil
}

// TouchLastSeen stores last-seen data.
func (repository *memoryUsers) TouchLastSeen(_ context.Context, id uuid.UUID) error {
	user, ok := repository.items[id]
	if !ok {
		return port.ErrNotFound
	}
	now := time.Now().UTC()
	user.LastSeenAt = &now
	repository.items[id] = user
	return nil
}

// memoryLinks stores identity links in memory.
type memoryLinks struct {
	items map[string]domain.IdentityLink
}

// Create stores an identity link.
func (repository *memoryLinks) Create(_ context.Context, link domain.IdentityLink) (domain.IdentityLink, error) {
	key := link.Issuer + ":" + link.Subject
	if _, ok := repository.items[key]; ok {
		return domain.IdentityLink{}, port.ErrConflict
	}
	repository.items[key] = link
	return link, nil
}

// FindByIssuerSubject returns one identity link.
func (repository *memoryLinks) FindByIssuerSubject(_ context.Context, issuer string, subject string) (domain.IdentityLink, error) {
	link, ok := repository.items[issuer+":"+subject]
	if !ok {
		return domain.IdentityLink{}, port.ErrNotFound
	}
	return link, nil
}

// Touch stores link touch data.
func (repository *memoryLinks) Touch(_ context.Context, link domain.IdentityLink) error {
	repository.items[link.Issuer+":"+link.Subject] = link
	return nil
}

// memoryClaims stores claim cache rows in memory.
type memoryClaims struct {
	items map[uuid.UUID]domain.ClaimCache
}

// Upsert stores claim cache data.
func (repository *memoryClaims) Upsert(_ context.Context, claims domain.ClaimCache) (domain.ClaimCache, error) {
	repository.items[claims.UserID] = claims
	return claims, nil
}

// FindByUserID returns claim cache data.
func (repository *memoryClaims) FindByUserID(_ context.Context, userID uuid.UUID) (domain.ClaimCache, error) {
	claims, ok := repository.items[userID]
	if !ok {
		return domain.ClaimCache{}, port.ErrNotFound
	}
	return claims, nil
}
