package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

// TestNewWritesJSON verifies logger output is valid structured JSON.
func TestNewWritesJSON(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Config{Level: "debug"}, WithOutput(&output))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Info("realmkit started", zap.String("component", "test"))
	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if entry["level"] != "info" {
		t.Fatalf("level = %v, want %v", entry["level"], "info")
	}
	if entry["message"] != "realmkit started" {
		t.Fatalf("message = %v, want %v", entry["message"], "realmkit started")
	}
	if entry["component"] != "test" {
		t.Fatalf("component = %v, want %v", entry["component"], "test")
	}
}

// TestNewHonorsConfiguredLevel verifies lower-severity entries are filtered.
func TestNewHonorsConfiguredLevel(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Config{Level: "warn"}, WithOutput(&output))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Info("hidden")
	log.Warn("visible")
	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if bytes.Contains(output.Bytes(), []byte("hidden")) {
		t.Fatalf("output contains filtered log entry: %s", output.String())
	}
	if !bytes.Contains(output.Bytes(), []byte("visible")) {
		t.Fatalf("output = %q, want visible entry", output.String())
	}
}

// TestNewDefaultsEmptyLevel verifies an empty log level falls back to info.
func TestNewDefaultsEmptyLevel(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Config{}, WithOutput(&output))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Debug("hidden")
	log.Info("visible")
	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if bytes.Contains(output.Bytes(), []byte("hidden")) {
		t.Fatalf("output contains debug entry: %s", output.String())
	}
	if !bytes.Contains(output.Bytes(), []byte("visible")) {
		t.Fatalf("output = %q, want info entry", output.String())
	}
}

// TestNewAcceptsErrorOutput verifies Zap internal error output can be redirected.
func TestNewAcceptsErrorOutput(t *testing.T) {
	var output bytes.Buffer
	var errorOutput bytes.Buffer
	if _, err := New(Config{Level: "info"}, WithOutput(&output), WithErrorOutput(&errorOutput)); err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

// TestNewRejectsInvalidLevel verifies invalid log levels fail fast.
func TestNewRejectsInvalidLevel(t *testing.T) {
	if _, err := New(Config{Level: "loud"}); err == nil {
		t.Fatalf("New() error = nil, want error")
	}
}
