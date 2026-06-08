package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub/backend/pkg/config"
	"github.com/niflaot/gamehub/backend/pkg/logger"
	"github.com/niflaot/gamehub/backend/pkg/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type failingSyncer struct {
	bytes.Buffer
}

func (syncer *failingSyncer) Sync() error {
	return errors.New("sync failed")
}

// TestRunReturnsConfiguredLoggerErrors verifies run returns configured logger errors explicitly.
func TestRunReturnsConfiguredLoggerErrors(t *testing.T) {
	t.Setenv("GAMEHUB_LOG_LEVEL", "loud")

	activeLogger := zap.NewNop()
	err := run(&activeLogger)
	if err == nil {
		t.Fatalf("run() error = nil, want error")
	}
}

// TestRunWithReturnsConfigErrors verifies startup stops when configuration fails.
func TestRunWithReturnsConfigErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("load failed")

	err := runWith(
		&activeLogger,
		func() (config.Config, error) {
			return config.Config{}, want
		},
		func(logger.Config) (*zap.Logger, error) {
			t.Fatalf("newLogger called after config failure")
			return nil, nil
		},
		func(*zap.Logger, bool) *fiber.App {
			t.Fatalf("newServer called after config failure")
			return nil
		},
		func(*fiber.App, string) error {
			t.Fatalf("listenServer called after config failure")
			return nil
		},
	)

	if !errors.Is(err, want) {
		t.Fatalf("runWith() error = %v, want %v", err, want)
	}
}

// TestRunWithReturnsLoggerErrors verifies startup stops when logger creation fails.
func TestRunWithReturnsLoggerErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("logger failed")

	err := runWith(
		&activeLogger,
		func() (config.Config, error) {
			return config.Config{}, nil
		},
		func(logger.Config) (*zap.Logger, error) {
			return nil, want
		},
		func(*zap.Logger, bool) *fiber.App {
			t.Fatalf("newServer called after logger failure")
			return nil
		},
		func(*fiber.App, string) error {
			t.Fatalf("listenServer called after logger failure")
			return nil
		},
	)

	if !errors.Is(err, want) {
		t.Fatalf("runWith() error = %v, want %v", err, want)
	}
}

// TestRunWithLogsStartup verifies startup logging uses Zap in every environment.
func TestRunWithLogsStartup(t *testing.T) {
	var output bytes.Buffer
	activeLogger := zap.NewNop()
	cfg := config.Config{
		Server:  server.Config{Host: "127.0.0.1", Port: 9090},
		Runtime: config.Runtime{Environment: "development"},
		Logging: logger.Config{Level: "info"},
	}

	err := runWith(
		&activeLogger,
		func() (config.Config, error) {
			return cfg, nil
		},
		func(cfg logger.Config) (*zap.Logger, error) {
			return logger.New(cfg, logger.WithOutput(&output))
		},
		func(_ *zap.Logger, development bool) *fiber.App {
			if !development {
				t.Fatalf("development = false, want true")
			}
			return fiber.New()
		},
		func(_ *fiber.App, address string) error {
			if address != "127.0.0.1:9090" {
				t.Fatalf("address = %q, want %q", address, "127.0.0.1:9090")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("runWith() error = %v", err)
	}

	if !bytes.Contains(output.Bytes(), []byte("starting gamehub backend")) {
		t.Fatalf("output = %q, want startup log", output.String())
	}
	if !bytes.Contains(output.Bytes(), []byte("127.0.0.1:9090")) {
		t.Fatalf("output = %q, want startup address", output.String())
	}
}

// TestRunWithLogsStartupOutsideDevelopment verifies startup logging is not environment-gated.
func TestRunWithLogsStartupOutsideDevelopment(t *testing.T) {
	var output bytes.Buffer
	activeLogger := zap.NewNop()
	cfg := config.Config{
		Server:  server.Config{Host: "127.0.0.1", Port: 9090},
		Runtime: config.Runtime{Environment: "production"},
		Logging: logger.Config{Level: "info"},
	}

	err := runWith(
		&activeLogger,
		func() (config.Config, error) {
			return cfg, nil
		},
		func(cfg logger.Config) (*zap.Logger, error) {
			return logger.New(cfg, logger.WithOutput(&output))
		},
		func(_ *zap.Logger, development bool) *fiber.App {
			if development {
				t.Fatalf("development = true, want false")
			}
			return fiber.New()
		},
		func(*fiber.App, string) error {
			return nil
		},
	)
	if err != nil {
		t.Fatalf("runWith() error = %v", err)
	}

	if !bytes.Contains(output.Bytes(), []byte("starting gamehub backend")) {
		t.Fatalf("output = %q, want startup log", output.String())
	}
}

// TestRunWithReturnsListenErrors verifies listener failures return explicitly.
func TestRunWithReturnsListenErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("listen failed")

	err := runWith(
		&activeLogger,
		func() (config.Config, error) {
			return config.Config{Logging: logger.Config{Level: "info"}}, nil
		},
		func(cfg logger.Config) (*zap.Logger, error) {
			return logger.New(cfg)
		},
		func(*zap.Logger, bool) *fiber.App {
			return fiber.New()
		},
		func(*fiber.App, string) error {
			return want
		},
	)

	if !errors.Is(err, want) {
		t.Fatalf("runWith() error = %v, want %v", err, want)
	}
}

// TestFinishLogsAndExitsOnError verifies finalization logs failures once and exits nonzero.
func TestFinishLogsAndExitsOnError(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(logger.Config{Level: "info"}, logger.WithOutput(&output))
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	code := 0
	finish(log, errors.New("boom"), func(value int) {
		code = value
	})

	if code != 1 {
		t.Fatalf("exit code = %d, want %d", code, 1)
	}
	if !bytes.Contains(output.Bytes(), []byte("gamehub backend failed")) {
		t.Fatalf("output = %q, want failure log", output.String())
	}
	if !bytes.Contains(output.Bytes(), []byte("boom")) {
		t.Fatalf("output = %q, want error details", output.String())
	}
}

// TestFinishDoesNotExitWithoutError verifies finalization exits only when an error exists.
func TestFinishDoesNotExitWithoutError(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(logger.Config{Level: "info"}, logger.WithOutput(&output))
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	called := false
	finish(log, nil, func(int) {
		called = true
	})

	if called {
		t.Fatalf("exit called = %v, want false", called)
	}
	if output.Len() != 0 {
		t.Fatalf("output = %q, want empty", output.String())
	}
}

// TestFinishExitsOnSyncError verifies logger sync failures are treated as finalization errors.
func TestFinishExitsOnSyncError(t *testing.T) {
	syncer := &failingSyncer{}
	log := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		syncer,
		zapcore.InfoLevel,
	))

	code := 0
	finish(log, nil, func(value int) {
		code = value
	})

	if code != 1 {
		t.Fatalf("exit code = %d, want %d", code, 1)
	}
	if !bytes.Contains(syncer.Bytes(), []byte("sync failed")) {
		t.Fatalf("output = %q, want sync error log", syncer.String())
	}
}

// TestSyncLoggerAcceptsNil verifies nil loggers can be finalized.
func TestSyncLoggerAcceptsNil(t *testing.T) {
	if err := syncLogger(nil); err != nil {
		t.Fatalf("syncLogger() error = %v", err)
	}
}
