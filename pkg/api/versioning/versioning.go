package versioning

import "github.com/gofiber/fiber/v2"

// APIPrefix is the root prefix for public GameHub APIs.
const APIPrefix = "/api"

// Version defines one public API version.
type Version struct {
	// Name is the version segment without a slash.
	Name string

	// Prefix is the absolute path prefix for the version.
	Prefix string
}

// V1 is the current stable public API version.
var V1 = Version{Name: "v1", Prefix: APIPrefix + "/v1"}

// Group returns a Fiber router scoped to the version prefix.
func (version Version) Group(app *fiber.App) fiber.Router {
	return app.Group(version.Prefix)
}
