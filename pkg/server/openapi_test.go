package server

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/openapi"
	"github.com/realmkit/rk-backend/pkg/api/swagger"
)

// TestRegisteredPublicRoutesExistInOpenAPI verifies Fiber routes are documented.
func TestRegisteredPublicRoutesExistInOpenAPI(t *testing.T) {
	app := newApp(t, nil, true)
	assertRoutesHaveOpenAPI(t, app)
}

// TestRegisteredMetadataRoutesExistInOpenAPI verifies optional metadata routes are documented.
func TestRegisteredMetadataRoutesExistInOpenAPI(t *testing.T) {
	app := newApp(t, nil, true, WithMetadata(newMetadataServices(t)))
	assertRoutesHaveOpenAPI(t, app)
}

// TestRegisteredGroupsRoutesExistInOpenAPI verifies optional groups routes are documented.
func TestRegisteredGroupsRoutesExistInOpenAPI(t *testing.T) {
	app := newApp(t, nil, true, WithGroups(newGroupsServices(t)))
	assertRoutesHaveOpenAPI(t, app)
}

// TestRegisteredUserRoutesExistInOpenAPI verifies optional user routes are documented.
func TestRegisteredUserRoutesExistInOpenAPI(t *testing.T) {
	authConfig, userService, userServices := newUserServices(t)
	app := newApp(t, nil, true, WithAuth(authConfig, userService), WithUsers(userServices))
	assertRoutesHaveOpenAPI(t, app)
}

// assertRoutesHaveOpenAPI verifies route contracts.
func assertRoutesHaveOpenAPI(t *testing.T, app *fiber.App) {
	t.Helper()
	for _, route := range app.GetRoutes() {
		if !requiresContract(route) {
			continue
		}
		ok, err := openapi.OperationExists(route.Method, route.Path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !ok {
			t.Fatalf("%s %s missing OpenAPI operation", route.Method, route.Path)
		}
	}
}

// requiresContract reports whether route must exist in OpenAPI.
func requiresContract(route fiber.Route) bool {
	if route.Method == fiber.MethodHead {
		return false
	}
	if route.Path == "/" && route.Method != fiber.MethodGet {
		return false
	}
	if route.Path == "/users" || route.Path == "/users/" {
		return false
	}
	if route.Path == swagger.DocsPath || route.Path == swagger.OpenAPIPath {
		return false
	}
	return route.Method != "USE"
}
