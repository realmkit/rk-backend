package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/events/adapter/postgres"
	"github.com/niflaot/gamehub-go/pkg/events/application"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
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
	resp, err := app.Test(newRequest(t, http.MethodGet, "/events/nope"))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
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
