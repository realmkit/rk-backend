package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	cronpostgres "github.com/realmkit/rk-backend/pkg/cronjob/adapter/postgres"
	cronapp "github.com/realmkit/rk-backend/pkg/cronjob/application"
	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/cronjob/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCronHTTPPauseAndResumeSuccess verifies state change routes with If-Match.
func TestCronHTTPPauseAndResumeSuccess(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	useTestPrincipal(app)
	Register(app, Services{Cron: newHTTPCronService(t), Checker: allowChecker{}})
	for _, route := range []struct {
		path    string
		ifMatch string
	}{
		{"/cronjobs/events.dispatch-pending/pause", `"1"`},
		{"/cronjobs/events.dispatch-pending/resume", `"2"`},
	} {
		req := newRequest(t, http.MethodPost, route.path)
		req.Header.Set(headers.IfMatch, route.ifMatch)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s error = %v", route.path, err)
		}
		if resp.StatusCode != fiber.StatusNoContent {
			t.Fatalf("%s status = %d, want 204", route.path, resp.StatusCode)
		}
	}
}

// TestCronHTTPProblemMappings covers adapter error responses.
func TestCronHTTPProblemMappings(t *testing.T) {
	cases := []struct {
		name    string
		service cronapp.Service
		method  string
		path    string
		ifMatch string
		status  int
	}{
		{
			"invalid pagination",
			newHTTPCronService(t),
			http.MethodGet,
			"/cronjobs?page_size=-1",
			"",
			fiber.StatusBadRequest,
		},
		{
			"invalid if match",
			newHTTPCronService(t),
			http.MethodPost,
			"/cronjobs/events.dispatch-pending/pause",
			"not-a-version",
			fiber.StatusBadRequest,
		},
		{
			"not found",
			newHTTPCronService(t),
			http.MethodGet,
			"/cronjobs/missing-job",
			"",
			fiber.StatusNotFound,
		},
		{
			"stale version",
			newHTTPCronService(t),
			http.MethodPost,
			"/cronjobs/events.dispatch-pending/pause",
			`"999"`,
			fiber.StatusPreconditionFailed,
		},
		{
			"handler missing",
			cronServiceWithoutHandlers(t),
			http.MethodPost,
			"/cronjobs/events.dispatch-pending/run",
			"",
			fiber.StatusConflict,
		},
	}

	for _, item := range cases {
		app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
		useTestPrincipal(app)
		Register(app, Services{Cron: item.service, Checker: allowChecker{}})
		req := newRequest(t, item.method, item.path)
		if item.ifMatch != "" {
			req.Header.Set(headers.IfMatch, item.ifMatch)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s error = %v", item.name, err)
		}
		if resp.StatusCode != item.status {
			t.Fatalf("%s status = %d, want %d", item.name, resp.StatusCode, item.status)
		}
	}
}

// cronServiceWithoutHandlers creates a due job service with no registered handler.
func cronServiceWithoutHandlers(t *testing.T) cronapp.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"-missing?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&cronpostgres.DefinitionModel{}, &cronpostgres.RunModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := cronpostgres.NewRepository(orm.NewStore(db))
	service := cronapp.NewService(
		cronapp.Dependencies{Repository: repo, Clock: cronHTTPClock{}},
		map[string]port.Handler{},
	)
	if err := service.EnsureDefinitions(context.Background(), []domain.Definition{cronHTTPDefinition()}); err != nil {
		t.Fatalf("EnsureDefinitions() error = %v", err)
	}
	return service
}
