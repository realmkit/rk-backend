package versioning

import "github.com/gofiber/fiber/v2"

// ServicePrefix is the unversioned in-service API prefix.
const ServicePrefix = ""

// Surface defines the route surface exposed by this service.
type Surface struct {
	// Name identifies who owns the public version boundary.
	Name string

	// Prefix is the absolute in-service path prefix.
	Prefix string
}

// Service is the GameHub route surface behind the API gateway.
var Service = Surface{Name: "gateway-owned", Prefix: ServicePrefix}

// Group returns a Fiber router scoped to the service prefix.
func (surface Surface) Group(app *fiber.App) fiber.Router {
	return app.Group(surface.Prefix)
}
