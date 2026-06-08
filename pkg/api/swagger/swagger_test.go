package swagger

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestRegisterSkipsRoutesWhenDisabled verifies Swagger routes are development-gated.
func TestRegisterSkipsRoutesWhenDisabled(t *testing.T) {
	app := fiber.New()
	Register(app, false)

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, DocsPath, nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNotFound)
	}
}

// TestRegisterServesOpenAPI verifies the raw OpenAPI document is served.
func TestRegisterServesOpenAPI(t *testing.T) {
	app := fiber.New()
	Register(app, true)

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, OpenAPIPath, nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusOK)
	}
	if !bytes.Contains(readBody(t, res), []byte(`"openapi"`)) {
		t.Fatalf("OpenAPI response missing openapi field")
	}
}

// TestRegisterServesUI verifies Swagger UI is served.
func TestRegisterServesUI(t *testing.T) {
	app := fiber.New()
	Register(app, true)

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, DocsPath, nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if !bytes.Contains(readBody(t, res), []byte("SwaggerUIBundle")) {
		t.Fatalf("Swagger UI response missing bundle")
	}
}

// readBody reads a response body for tests.
func readBody(t *testing.T, res *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return body
}
