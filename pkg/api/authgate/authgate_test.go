package authgate

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/principal"
)

// TestRequirePrincipalMapsMissingPrincipal verifies missing auth fails closed.
func TestRequirePrincipalMapsMissingPrincipal(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		_, err := RequirePrincipal(ctx)
		if err == nil {
			t.Fatalf("RequirePrincipal() error = nil, want error")
		}
		return nil
	})
	if _, err := app.Test(httptestRequest(t)); err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

// TestUserIDHelpersUsePrincipal verifies required and optional user extraction.
func TestUserIDHelpersUsePrincipal(t *testing.T) {
	userID := uuid.New()
	app := fiber.New()
	app.Use(func(ctx *fiber.Ctx) error {
		principal.Set(ctx, principal.Principal{UserID: userID, Scopes: []string{"read:account"}})
		return ctx.Next()
	})
	app.Get("/", func(ctx *fiber.Ctx) error {
		required, err := RequireUserID(ctx)
		if err != nil {
			t.Fatalf("RequireUserID() error = %v", err)
		}
		if required != userID || OptionalUserID(ctx) != userID {
			t.Fatalf("user IDs = %s %s, want %s", required, OptionalUserID(ctx), userID)
		}
		return nil
	})
	if _, err := app.Test(httptestRequest(t)); err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

// TestScopeGatesVerifyPrincipalScopes verifies scope allow and deny behavior.
func TestScopeGatesVerifyPrincipalScopes(t *testing.T) {
	app := fiber.New()
	app.Use(func(ctx *fiber.Ctx) error {
		principal.Set(ctx, principal.Principal{UserID: uuid.New(), Scopes: []string{"admin:realmkit"}})
		return ctx.Next()
	})
	app.Get("/", func(ctx *fiber.Ctx) error {
		if err := RequireScope(ctx, "admin:realmkit"); err != nil {
			t.Fatalf("RequireScope(admin) error = %v", err)
		}
		if err := RequireAnyScope(ctx, "missing", "other"); err == nil {
			t.Fatalf("RequireAnyScope(missing) error = nil, want error")
		}
		return nil
	})
	if _, err := app.Test(httptestRequest(t)); err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
}

// httptestRequest creates a basic Fiber request.
func httptestRequest(t *testing.T) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	return request
}
