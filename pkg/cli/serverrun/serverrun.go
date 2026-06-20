package serverrun

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/server"
)

// Serve starts app and shuts it down when ctx is canceled.
func Serve(ctx context.Context, app *fiber.App, address string, cfg server.Config) error {
	errs := make(chan error, 1)
	go func() {
		errs <- app.Listen(address)
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.Defaults().ShutdownTimeout)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			return err
		}
		return <-errs
	}
}
