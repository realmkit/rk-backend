// Package cronjob_e2e verifies cronjob runtime behavior.
package cronjob_e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
	cronhttp "github.com/niflaot/gamehub-go/pkg/cronjob/adapter/http"
	cronpostgres "github.com/niflaot/gamehub-go/pkg/cronjob/adapter/postgres"
	cronapp "github.com/niflaot/gamehub-go/pkg/cronjob/application"
	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/cronjob/port"
	"github.com/niflaot/gamehub-go/pkg/server"
)

// TestCronjobManualRunThroughHTTP verifies cron HTTP execution.
func TestCronjobManualRunThroughHTTP(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("create shared migrated database and cron service")
	database := harness.NewSQLiteDatabase(t)
	repository := cronpostgres.NewRepository(database.Store)
	service := cronapp.NewService(
		cronapp.Dependencies{Repository: repository, WorkerID: "e2e-worker"},
		map[string]port.Handler{
			"e2e.job": port.HandlerFunc(func(context.Context, port.RunContext) (domain.Result, error) {
				return domain.Result{ProcessedCount: 1, ChangedCount: 1}, nil
			}),
		},
	)
	if err := service.EnsureDefinitions(context.Background(), []domain.Definition{cronDefinition()}); err != nil {
		t.Fatalf("EnsureDefinitions() error = %v", err)
	}

	steps.Log("start server with cron routes")
	ecosystem := harness.New(
		t,
		harness.WithDatabase(database),
		harness.WithServerOptions(server.WithCron(cronhttp.Services{Cron: service})),
	)

	steps.Log("trigger cron job through HTTP")
	request := harness.JSONRequest(fiber.MethodPost, "/cronjobs/e2e.job/run", "")
	response := ecosystem.Test(t, request)
	body := harness.ResponseBody(t, response)
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, fiber.StatusOK, body)
	}

	steps.Log("verify run history through HTTP")
	response = ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/cronjobs/e2e.job/runs?page_size=5", ""))
	body = harness.ResponseBody(t, response)
	if response.StatusCode != fiber.StatusOK || !strings.Contains(body, "processed_count") {
		t.Fatalf("runs response status=%d body=%q, want processed history", response.StatusCode, body)
	}
}

// cronDefinition returns the e2e manual cron job definition.
func cronDefinition() domain.Definition {
	return domain.Definition{
		Key:               "e2e.job",
		Name:              "E2E Job",
		Description:       "Runs during e2e tests.",
		ScheduleKind:      domain.ScheduleManual,
		Enabled:           true,
		ConcurrencyPolicy: domain.ConcurrencyForbid,
	}
}
