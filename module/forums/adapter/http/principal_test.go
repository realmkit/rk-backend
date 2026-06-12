package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/api/problem"
)

// currentUserIDHeader is a test-only identity header.
const currentUserIDHeader = "X-RealmKit-Test-User-Id"

// useTestPrincipal installs test-only principal injection.
func useTestPrincipal(app *fiber.App) {
	app.Use(func(ctx *fiber.Ctx) error {
		value := strings.TrimSpace(ctx.Get(currentUserIDHeader))
		if value == "" {
			return ctx.Next()
		}
		userID, err := uuid.Parse(value)
		if err != nil {
			return problem.Write(ctx, problem.New(fiber.StatusBadRequest, "invalid_test_user", currentUserIDHeader+" must be a UUID."))
		}
		principal.Set(ctx, principal.Principal{UserID: userID, SubjectHash: "test:" + userID.String()})
		return ctx.Next()
	})
}
