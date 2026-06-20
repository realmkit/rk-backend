package server

import (
	"testing"
	"time"
)

// TestConfigAddress verifies server addresses are formatted for Listen.
func TestConfigAddress(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 9090}

	if cfg.Address() != "127.0.0.1:9090" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "127.0.0.1:9090")
	}
}

// TestConfigDefaults verifies server runtime defaults are applied.
func TestConfigDefaults(t *testing.T) {
	cfg := Config{}.Defaults()

	if cfg.RequestTimeout != 15*time.Second {
		t.Fatalf("RequestTimeout = %s, want 15s", cfg.RequestTimeout)
	}
	if cfg.UploadRequestTimeout != 120*time.Second {
		t.Fatalf("UploadRequestTimeout = %s, want 120s", cfg.UploadRequestTimeout)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 15s", cfg.ShutdownTimeout)
	}
}

// TestNewConfiguresTimeouts verifies Fiber IO timeouts come from server config.
func TestNewConfiguresTimeouts(t *testing.T) {
	cfg := Config{
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   3 * time.Second,
		IdleTimeout:    4 * time.Second,
		RequestTimeout: 5 * time.Second,
	}
	app := newApp(t, nil, false, WithConfig(cfg))

	if app.Config().ReadTimeout != 2*time.Second {
		t.Fatalf("ReadTimeout = %s, want 2s", app.Config().ReadTimeout)
	}
	if app.Config().WriteTimeout != 3*time.Second {
		t.Fatalf("WriteTimeout = %s, want 3s", app.Config().WriteTimeout)
	}
	if app.Config().IdleTimeout != 4*time.Second {
		t.Fatalf("IdleTimeout = %s, want 4s", app.Config().IdleTimeout)
	}
}
