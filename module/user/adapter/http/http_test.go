package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/user/domain"
	userport "github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/principal"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
)

// TestCurrentUserRouteReturnsUser verifies current user response.
func TestCurrentUserRouteReturnsUser(t *testing.T) {
	userID := uuid.New()
	app := testApp(userID, &userService{current: testCurrentUser(userID)})
	res, err := app.Test(testRequest(http.MethodGet, "/users/me", ""))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusOK || res.Header.Get(headers.ETag) == "" {
		t.Fatalf("status=%d etag=%q, want current user", res.StatusCode, res.Header.Get(headers.ETag))
	}
}

// TestUpdateCurrentUserRequiresIdempotency verifies update headers.
func TestUpdateCurrentUserRequiresIdempotency(t *testing.T) {
	userID := uuid.New()
	app := testApp(
		userID,
		&userService{user: domain.User{ID: userID, Status: domain.StatusActive, FirstSeenAt: time.Now().UTC(), Version: 1}},
	)
	res, err := app.Test(testRequest(http.MethodPatch, "/users/me", `{}`))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want 400", res.StatusCode)
	}
}

// TestUpdateCurrentUserReturnsUpdatedUser verifies update success.
func TestUpdateCurrentUserReturnsUpdatedUser(t *testing.T) {
	userID := uuid.New()
	avatarID := uuid.New()
	app := testApp(
		userID,
		&userService{
			user: domain.User{ID: userID, Status: domain.StatusActive, AvatarAssetID: &avatarID, FirstSeenAt: time.Now().UTC(), Version: 2},
		},
	)
	req := testRequest(http.MethodPatch, "/users/me", `{"avatar_asset_id":"`+avatarID.String()+`"}`)
	req.Header.Set(headers.IdempotencyKey, "update")
	req.Header.Set(headers.IfMatch, `"1"`)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("StatusCode = %d body=%s, want 200", res.StatusCode, body)
	}
}

// TestAccountURLUnavailableMapsProblem verifies optional account URL response.
func TestAccountURLUnavailableMapsProblem(t *testing.T) {
	userID := uuid.New()
	app := testApp(userID, &userService{current: testCurrentUser(userID)})
	res, err := app.Test(testRequest(http.MethodGet, "/users/me/identity/account-url", ""))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want 404", res.StatusCode)
	}
}

// TestCurrentUserMapsNotFound verifies not-found problem mapping.
func TestCurrentUserMapsNotFound(t *testing.T) {
	userID := uuid.New()
	app := testApp(userID, &userService{err: userport.ErrNotFound})
	res, err := app.Test(testRequest(http.MethodGet, "/users/me", ""))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want 404", res.StatusCode)
	}
}

// TestUpdateCurrentUserRejectsInvalidJSON verifies strict JSON decoding.
func TestUpdateCurrentUserRejectsInvalidJSON(t *testing.T) {
	userID := uuid.New()
	app := testApp(
		userID,
		&userService{user: domain.User{ID: userID, Status: domain.StatusActive, FirstSeenAt: time.Now().UTC(), Version: 1}},
	)
	req := testRequest(http.MethodPatch, "/users/me", `{"avatar_asset_id":`)
	req.Header.Set(headers.IdempotencyKey, "update")
	req.Header.Set(headers.IfMatch, `"1"`)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want 400", res.StatusCode)
	}
}

// testApp creates a user HTTP app.
func testApp(userID uuid.UUID, service *userService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	authenticate := func(ctx *fiber.Ctx) error {
		principal.Set(ctx, principal.Principal{UserID: userID, Issuer: "test", SubjectHash: "hash"})
		return ctx.Next()
	}
	Register(app, Services{Users: service}, authenticate)
	return app
}

// testRequest creates a JSON request.
func testRequest(method string, target string, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set(headers.Accept, "application/json")
	if body != "" {
		req.Header.Set(headers.ContentType, "application/json")
	}
	return req
}

// testCurrentUser returns a current user response.
func testCurrentUser(userID uuid.UUID) userport.CurrentUser {
	now := time.Now().UTC()
	claims := domain.ClaimCache{ID: uuid.New(), UserID: userID, Username: "ian", ClaimsHash: "hash", SyncedAt: now}
	return userport.CurrentUser{User: domain.User{ID: userID, Status: domain.StatusActive, FirstSeenAt: now, Version: 1}, Claims: &claims}
}

// userService is a fake user service.
type userService struct {
	current userport.CurrentUser
	user    domain.User
	err     error
}

// Get returns one user.
func (service *userService) Get(context.Context, uuid.UUID) (domain.User, error) {
	return service.user, service.err
}

// Current returns current user data.
func (service *userService) Current(context.Context, uuid.UUID) (userport.CurrentUser, error) {
	return service.current, service.err
}

// UpdateCurrent updates current user data.
func (service *userService) UpdateCurrent(context.Context, userport.UpdateCurrentCommand) (domain.User, error) {
	return service.user, service.err
}
