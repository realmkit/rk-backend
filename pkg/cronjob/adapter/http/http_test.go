package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	cronpostgres "github.com/niflaot/gamehub-go/pkg/cronjob/adapter/postgres"
	cronapp "github.com/niflaot/gamehub-go/pkg/cronjob/application"
	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/cronjob/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCronHTTPRoutes verifies cron admin route behavior.
func TestCronHTTPRoutes(t *testing.T) {
	service := newHTTPCronService(t)
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{Cron: service})

	for _, req := range []*http.Request{
		newRequest(t, http.MethodGet, "/cronjobs"),
		newRequest(t, http.MethodGet, "/cronjobs/events.dispatch-pending"),
		newRequest(t, http.MethodGet, "/cronjobs/events.dispatch-pending/runs"),
		newRequest(t, http.MethodPost, "/cronjobs/events.dispatch-pending/run"),
		newRequest(t, http.MethodPost, "/cronjobs/locks/repair"),
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

// TestCronHTTPPauseRequiresIfMatch verifies concurrency header handling.
func TestCronHTTPPauseRequiresIfMatch(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{Cron: newHTTPCronService(t)})
	resp, err := app.Test(newRequest(t, http.MethodPost, "/cronjobs/events.dispatch-pending/pause"))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusPreconditionRequired {
		t.Fatalf("status = %d, want 428", resp.StatusCode)
	}
}

// newHTTPCronService creates a SQLite-backed cron service.
func newHTTPCronService(t *testing.T) cronapp.Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&cronpostgres.DefinitionModel{}, &cronpostgres.RunModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	repo := cronpostgres.NewRepository(orm.NewStore(db))
	service := cronapp.NewService(
		cronapp.Dependencies{Repository: repo, Clock: cronHTTPClock{}},
		map[string]port.Handler{domain.JobEventsDispatchPending: cronHTTPHandler{}},
	)
	if err := service.EnsureDefinitions(context.Background(), []domain.Definition{cronHTTPDefinition()}); err != nil {
		t.Fatalf("EnsureDefinitions() error = %v", err)
	}
	return service
}

// cronHTTPHandler is a successful test handler.
type cronHTTPHandler struct{}

// Run returns one processed item.
func (cronHTTPHandler) Run(context.Context, port.RunContext) (domain.Result, error) {
	return domain.Result{ProcessedCount: 1}, nil
}

// cronHTTPClock returns fixed time.
type cronHTTPClock struct{}

// Now returns fixed time.
func (cronHTTPClock) Now() time.Time {
	return time.Unix(500, 0).UTC()
}

// cronHTTPDefinition returns a due job definition.
func cronHTTPDefinition() domain.Definition {
	now := time.Unix(500, 0).UTC()
	return domain.Definition{
		Key:                domain.JobEventsDispatchPending,
		Name:               "Dispatch events",
		ScheduleKind:       domain.ScheduleInterval,
		ScheduleExpression: time.Minute.String(),
		Enabled:            true,
		ConcurrencyPolicy:  domain.ConcurrencyForbid,
		NextRunAt:          &now,
		Version:            1,
	}
}

// newRequest creates a request.
func newRequest(t *testing.T, method string, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	return req
}
