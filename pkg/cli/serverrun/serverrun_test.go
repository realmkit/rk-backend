package serverrun

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/server"
)

// TestServeShutsDownWhenContextCancels verifies Serve honors root cancellation.
func TestServeShutsDownWhenContextCancels(t *testing.T) {
	app := fiber.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := Serve(ctx, app, "127.0.0.1:0", server.Config{ShutdownTimeout: time.Second})
	if err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
}
