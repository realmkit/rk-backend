package principal

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TestContextStoresPrincipal verifies context helpers.
func TestContextStoresPrincipal(t *testing.T) {
	expected := Principal{UserID: uuid.New(), Issuer: "issuer", SubjectHash: "hash"}
	ctx := WithContext(context.Background(), expected)
	got, ok := FromContext(ctx)
	if !ok || got.UserID != expected.UserID {
		t.Fatalf("FromContext() = %+v, %v, want principal", got, ok)
	}
}

// TestFiberLocalsStorePrincipal verifies Fiber helper behavior.
func TestFiberLocalsStorePrincipal(t *testing.T) {
	app := fiber.New()
	expected := Principal{UserID: uuid.New(), Issuer: "issuer", SubjectHash: "hash"}
	app.Get("/", func(ctx *fiber.Ctx) error {
		Set(ctx, expected)
		got, err := Require(ctx)
		if err != nil {
			t.Fatalf("Require() error = %v", err)
		}
		if got.UserID != expected.UserID {
			t.Fatalf("Require() = %+v, want %+v", got, expected)
		}
		return nil
	})
	if _, err := app.Test(httptestRequest("GET", "/")); err != nil {
		t.Fatalf("Test() error = %v", err)
	}
}

// TestRequireMissingReturnsError verifies missing principal behavior.
func TestRequireMissingReturnsError(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		if _, err := Require(ctx); err != ErrMissing {
			t.Fatalf("Require() error = %v, want ErrMissing", err)
		}
		return nil
	})
	if _, err := app.Test(httptestRequest("GET", "/")); err != nil {
		t.Fatalf("Test() error = %v", err)
	}
}

// httptestRequest creates a request without importing net/http in every test.
func httptestRequest(method string, target string) *http.Request {
	request, _ := http.NewRequest(method, target, nil)
	return request
}
