package problem

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/headers"
)

// TestHandlerWritesFiberErrors verifies Fiber errors become problem responses.
func TestHandlerWritesFiberErrors(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: Handler})
	app.Use(headers.Middleware())
	app.Get("/missing", func(*fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "gone")
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/missing", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	var payload Problem
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if payload.Status != fiber.StatusNotFound {
		t.Fatalf("Status = %d, want %d", payload.Status, fiber.StatusNotFound)
	}
	if payload.Code != "not_found" {
		t.Fatalf("Code = %q, want %q", payload.Code, "not_found")
	}
}

// TestHandlerWritesProblemErrors verifies explicit problem errors are preserved.
func TestHandlerWritesProblemErrors(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: Handler})
	app.Get("/invalid", func(*fiber.Ctx) error {
		return Error{Problem: New(fiber.StatusConflict, "duplicate", "duplicate request")}
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/invalid", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	var payload Problem
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if payload.Code != "duplicate" {
		t.Fatalf("Code = %q, want %q", payload.Code, "duplicate")
	}
}

// TestErrorReturnsDetailOrTitle verifies problem errors prefer detail.
func TestErrorReturnsDetailOrTitle(t *testing.T) {
	withDetail := Error{Problem: New(fiber.StatusConflict, "duplicate", "duplicate request")}
	if withDetail.Error() != "duplicate request" {
		t.Fatalf("Error() = %q, want detail", withDetail.Error())
	}

	withoutDetail := Error{Problem: New(fiber.StatusTeapot, "teapot", "")}
	if withoutDetail.Error() != "I'm a teapot" {
		t.Fatalf("Error() = %q, want title", withoutDetail.Error())
	}
}

// TestHandlerWritesInternalErrors verifies ordinary errors become internal problems.
func TestHandlerWritesInternalErrors(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: Handler})
	app.Get("/boom", func(*fiber.Ctx) error {
		return errors.New("boom")
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/boom", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusInternalServerError)
	}
}
