package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/config"
	"github.com/niflaot/gamehub-go/pkg/logger"
	"github.com/niflaot/gamehub-go/pkg/postgres"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"github.com/niflaot/gamehub-go/pkg/server"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRootCommandShowsHelpByDefault verifies no-arg execution shows commands.
func TestRootCommandShowsHelpByDefault(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "start") || !strings.Contains(output.String(), "migrate") {
		t.Fatalf("output = %q, want start and migrate commands", output.String())
	}
}

// TestRootCommandHelpPrintsUsage verifies help is the usage-printing path.
func TestRootCommandHelpPrintsUsage(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want Usage", output.String())
	}
}

// TestRootCommandErrorDoesNotPrintUsage verifies errors do not include usage output.
func TestRootCommandErrorDoesNotPrintUsage(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "repair"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("Execute() error = nil, want error")
	}
	if strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want no Usage", output.String())
	}
}

// TestMigrateStatusReportsPendingMigration verifies migrate status uses the global runner.
func TestMigrateStatusReportsPendingMigration(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "pending=1") {
		t.Fatalf("output = %q, want pending=1", output.String())
	}
}

// TestMigrateDownRequiresDestructiveConfirmation verifies rollback is guarded.
func TestMigrateDownRequiresDestructiveConfirmation(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "down"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("Execute() error = nil, want error")
	}
	if strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want no Usage", output.String())
	}
}

// TestExecuteReturnsStartErrors verifies the start command serves the API.
func TestExecuteReturnsStartErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("listen failed")
	deps := testCommandDeps(t)
	deps.listenServer = func(*fiber.App, string) error {
		return want
	}

	err := execute(&activeLogger, []string{"start"}, deps)
	if !errors.Is(err, want) {
		t.Fatalf("execute() error = %v, want %v", err, want)
	}
}

// TestRunStartLogsStartup verifies startup logging uses Zap in every environment.
func TestRunStartLogsStartup(t *testing.T) {
	var output bytes.Buffer
	activeLogger := zap.NewNop()
	cfg := config.Config{
		Server:  server.Config{Host: "127.0.0.1", Port: 9090},
		Runtime: config.Runtime{Environment: "development"},
		Logging: logger.Config{Level: "info"},
	}
	deps := testCommandDeps(t)
	deps.loadConfig = func() (config.Config, error) {
		return cfg, nil
	}
	deps.newLogger = func(cfg logger.Config) (*zap.Logger, error) {
		return logger.New(cfg, logger.WithOutput(&output))
	}
	deps.newServer = func(_ *zap.Logger, development bool) *fiber.App {
		if !development {
			t.Fatalf("development = false, want true")
		}
		return fiber.New()
	}
	deps.listenServer = func(_ *fiber.App, address string) error {
		if address != "127.0.0.1:9090" {
			t.Fatalf("address = %q, want %q", address, "127.0.0.1:9090")
		}
		return nil
	}

	if err := execute(&activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if !bytes.Contains(output.Bytes(), []byte("starting gamehub backend")) {
		t.Fatalf("output = %q, want startup log", output.String())
	}
}

// testCommandDeps returns deterministic command dependencies.
func testCommandDeps(t *testing.T) commandDeps {
	t.Helper()
	return commandDeps{
		loadConfig: func() (config.Config, error) {
			return config.Config{Logging: logger.Config{Level: "info"}}, nil
		},
		newLogger: func(logger.Config) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		newServer: func(*zap.Logger, bool) *fiber.App {
			return fiber.New()
		},
		listenServer: func(*fiber.App, string) error {
			return nil
		},
		openPostgres: func(context.Context, postgres.Config) (*gorm.DB, error) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("gorm.Open() error = %v", err)
			}
			return db, nil
		},
		closePostgres: func(*gorm.DB) error {
			return nil
		},
		newRunner: func(db *gorm.DB, log *zap.Logger) migrations.Runner {
			return migrations.NewRunner(db, migrations.DefaultSource(), migrations.WithLogger(log))
		},
	}
}
