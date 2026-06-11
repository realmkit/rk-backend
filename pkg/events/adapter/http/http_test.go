package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/events/adapter/postgres"
	"github.com/niflaot/gamehub-go/pkg/events/application"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEventHTTPRoutes verifies admin event routes.
func TestEventHTTPRoutes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&postgres.EventModel{}, &postgres.ScopeModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := postgres.NewRepository(orm.NewStore(db))
	service := application.NewService(application.Dependencies{Repository: repo})
	event, err := service.Publish(context.Background(), httpDraft())
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{Events: service, Hub: NewHub()})

	for _, req := range []*http.Request{
		newRequest(t, http.MethodGet, "/events"),
		newRequest(t, http.MethodGet, "/events/"+event.ID.String()),
		newRequest(t, http.MethodPost, "/events/"+event.ID.String()+"/cancel"),
		newRequest(t, http.MethodPost, "/events/"+event.ID.String()+"/replay"),
	} {
		req.Header.Set(currentUserIDHeader, uuid.NewString())
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", req.Method, req.URL.Path, err)
		}
		if resp.StatusCode >= 400 {
			t.Fatalf("%s %s status = %d, want success", req.Method, req.URL.Path, resp.StatusCode)
		}
	}
}

// TestEventHTTPInvalidID verifies problem mapping.
func TestEventHTTPInvalidID(t *testing.T) {
	service := newHTTPEventService(t)
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{Events: service})
	req := newRequest(t, http.MethodGet, "/events/nope")
	req.Header.Set(currentUserIDHeader, uuid.NewString())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

// TestEventAdminRoutesRequireUser verifies event diagnostics are not public.
func TestEventAdminRoutesRequireUser(t *testing.T) {
	service := newHTTPEventService(t)
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{Events: service})
	id := uuid.NewString()
	for _, req := range []*http.Request{
		newRequest(t, http.MethodGet, "/events"),
		newRequest(t, http.MethodGet, "/events/"+id),
		newRequest(t, http.MethodPost, "/events/"+id+"/cancel"),
		newRequest(t, http.MethodPost, "/events/"+id+"/replay"),
	} {
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", req.Method, req.URL.Path, err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want 401", req.Method, req.URL.Path, resp.StatusCode)
		}
	}
}

// TestWebSocketScopeAuthorizationPreventsLeaks verifies private scopes fail closed.
func TestWebSocketScopeAuthorizationPreventsLeaks(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	privateScope := domain.Scope{Type: domain.ScopeTicket, ID: uuid.NewString()}
	client := &client{userID: userID, scopes: map[string]domain.Scope{}}
	for _, scope := range []domain.Scope{
		{Type: domain.ScopeUser, ID: otherUserID.String()},
		{Type: domain.ScopeStaff},
		{Type: domain.ScopePermission, Permission: "tickets.view"},
		{Type: domain.ScopeSystem},
		privateScope,
	} {
		if client.canSubscribe(scope) {
			t.Fatalf("canSubscribe(%+v) = true, want false", scope)
		}
	}
	client.authz = scopeAuthorizer{allowed: map[string]bool{scopeKey(privateScope): true}}
	if !client.canSubscribe(privateScope) {
		t.Fatalf("canSubscribe(ticket scope) = false, want true with authorizer grant")
	}
	if client.canSubscribe(domain.Scope{Type: domain.ScopeSystem}) {
		t.Fatalf("canSubscribe(system) = true, want false even with authorizer")
	}
}

// TestWebSocketScopeAuthorizationAllowsGlobalAndOwnUser verifies safe self-service subscriptions.
func TestWebSocketScopeAuthorizationAllowsGlobalAndOwnUser(t *testing.T) {
	userID := uuid.New()
	client := &client{userID: userID, scopes: map[string]domain.Scope{}}
	if !client.canSubscribe(domain.Scope{Type: domain.ScopeGlobal}) {
		t.Fatalf("canSubscribe(global) = false, want true")
	}
	if !client.canSubscribe(domain.Scope{Type: domain.ScopeUser, ID: userID.String()}) {
		t.Fatalf("canSubscribe(own user) = false, want true")
	}
}

// newHTTPEventService creates a SQLite-backed event service.
func newHTTPEventService(t *testing.T) application.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&postgres.EventModel{}, &postgres.ScopeModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return application.NewService(application.Dependencies{
		Repository: postgres.NewRepository(orm.NewStore(db)),
	})
}

// httpDraft returns a valid draft.
func httpDraft() domain.Draft {
	return domain.Draft{
		Key:           domain.EventForumsThreadCreated,
		SchemaVersion: 1,
		Producer:      domain.ProducerForums,
		AggregateType: "forum_thread",
		Payload:       map[string]any{},
		Scopes:        []domain.Scope{{Type: domain.ScopeStaff}},
	}
}

// newRequest creates an HTTP request.
func newRequest(t *testing.T, method string, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	return req
}

type scopeAuthorizer struct {
	allowed map[string]bool
}

func (authorizer scopeAuthorizer) CanSubscribe(
	_ context.Context,
	_ port.Principal,
	scope domain.Scope,
) (bool, error) {
	return authorizer.allowed[scopeKey(scope)], nil
}
