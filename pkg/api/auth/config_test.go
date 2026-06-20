package auth

import (
	"testing"
	"time"
)

// TestConfigDefaults verifies outbound auth HTTP calls have bounded defaults.
func TestConfigDefaults(t *testing.T) {
	cfg := Config{}.Defaults()
	if cfg.HTTPTimeout != 5*time.Second {
		t.Fatalf("HTTPTimeout = %s, want 5s", cfg.HTTPTimeout)
	}
	validator := NewValidator(Config{HTTPTimeout: time.Second})
	if validator.client.Timeout != time.Second {
		t.Fatalf("client timeout = %s, want 1s", validator.client.Timeout)
	}
}
