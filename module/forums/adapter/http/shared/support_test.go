package shared

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
)

// TestSupportHelpersCoverTransportBranches verifies shared HTTP helper behavior.
func TestSupportHelpersCoverTransportBranches(t *testing.T) {
	app := fiber.New()
	app.Post("/json", func(ctx *fiber.Ctx) error {
		var payload struct {
			Name string `json:"name"`
		}
		if err := DecodeJSON(ctx, &payload); err != nil {
			return err
		}
		return WriteJSON(ctx, fiber.StatusAccepted, payload)
	})
	app.Get("/auth/:id", func(ctx *fiber.Ctx) error {
		if _, err := IDFromParam(ctx, "id"); err != nil {
			return err
		}
		if _, err := CurrentUserID(ctx); err != nil {
			return err
		}
		SetETag(ctx, 7)
		return WriteNoContent(ctx)
	})
	app.Post("/headers", func(ctx *fiber.Ctx) error {
		if err := RequireIdempotency(ctx); err != nil {
			return err
		}
		if _, err := ExpectedVersion(ctx); err != nil {
			return err
		}
		if _, err := PageFromQuery(ctx); err != nil {
			return err
		}
		return WriteNoContent(ctx)
	})
	app.Get("/problem", func(ctx *fiber.Ctx) error {
		return HandleError(ctx, port.ErrForbidden)
	})
	app.Get("/passthrough", func(ctx *fiber.Ctx) error {
		return HandleError(ctx, errors.New("plain"))
	})

	assertStatus(t, app, http.MethodPost, "/json", `{"name":"ok"}`, nil, fiber.StatusAccepted)
	assertStatus(t, app, http.MethodPost, "/json", `{`, nil, fiber.StatusInternalServerError)
	assertStatus(t, app, http.MethodGet, "/auth/not-uuid", "", nil, fiber.StatusInternalServerError)
	assertStatus(t, app, http.MethodGet, "/auth/"+uuid.NewString(), "", nil, fiber.StatusInternalServerError)

	headers := map[string]string{CurrentUserIDHeader: uuid.NewString()}
	resp := assertStatus(t, app, http.MethodGet, "/auth/"+uuid.NewString(), "", headers, fiber.StatusNoContent)
	if resp.Header.Get(headerspkgETag()) != `"7"` {
		t.Fatalf("ETag = %q, want %q", resp.Header.Get(headerspkgETag()), `"7"`)
	}

	assertStatus(t, app, http.MethodPost, "/headers", "", nil, fiber.StatusInternalServerError)
	headers = map[string]string{
		headerspkgIdempotencyKey(): "key",
		headerspkgIfMatch():        `"1"`,
	}
	assertStatus(t, app, http.MethodPost, "/headers?page_size=1", "", headers, fiber.StatusNoContent)
	assertStatus(t, app, http.MethodGet, "/problem", "", nil, fiber.StatusForbidden)
	assertStatus(t, app, http.MethodGet, "/passthrough", "", nil, fiber.StatusInternalServerError)
}

// assertStatus performs one request and checks its status.
func assertStatus(
	t *testing.T,
	app *fiber.App,
	method string,
	path string,
	body string,
	values map[string]string,
	want int,
) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(method, path, stringsReader(body))
	req.Header.Set(headers.ContentType, "application/json")
	for key, value := range values {
		req.Header.Set(key, value)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s Test() error = %v", method, path, err)
	}
	if resp.StatusCode != want {
		t.Fatalf("%s %s status = %d, want %d", method, path, resp.StatusCode, want)
	}
	return resp
}

// stringsReader wraps a string as an HTTP body reader.
func stringsReader(value string) *strings.Reader {
	return strings.NewReader(value)
}

// headerspkgETag returns the ETag header name.
func headerspkgETag() string {
	return headers.ETag
}

// headerspkgIdempotencyKey returns the Idempotency-Key header name.
func headerspkgIdempotencyKey() string {
	return headers.IdempotencyKey
}

// headerspkgIfMatch returns the If-Match header name.
func headerspkgIfMatch() string {
	return headers.IfMatch
}
