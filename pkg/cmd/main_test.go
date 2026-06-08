package main

import (
	"bytes"
	"errors"
	"syscall"
	"testing"

	"github.com/niflaot/gamehub-go/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// failingSyncer captures logs and fails when Sync is called.
type failingSyncer struct {
	bytes.Buffer
}

// unsupportedSyncer captures logs and reports an unsupported stdout sync failure.
type unsupportedSyncer struct {
	bytes.Buffer
	err error
}

// Sync returns a deterministic sync failure for finalizer tests.
func (syncer *failingSyncer) Sync() error {
	return errors.New("sync failed")
}

// Sync returns an unsupported stdout sync failure.
func (syncer *unsupportedSyncer) Sync() error {
	return syncer.err
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

// TestSyncLoggerIgnoresUnsupportedOutputSync verifies stdout sync incompatibility is tolerated.
func TestSyncLoggerIgnoresUnsupportedOutputSync(t *testing.T) {
	cases := []error{syscall.EINVAL, syscall.EBADF, syscall.ENOTTY}

	for _, syncErr := range cases {
		log := zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			&unsupportedSyncer{err: syncErr},
			zapcore.InfoLevel,
		))

		if err := syncLogger(log); err != nil {
			t.Fatalf("syncLogger() error = %v", err)
		}
	}
}
