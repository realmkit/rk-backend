package swagger

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/openapi"
)

// DocsPath is the Swagger UI route.
const DocsPath = "/docs"

// OpenAPIPath is the raw OpenAPI document route.
const OpenAPIPath = "/docs/openapi.json"

// Register mounts Swagger routes when enabled is true.
func Register(app *fiber.App, enabled bool) {
	if !enabled {
		return
	}

	app.Get(OpenAPIPath, document)
	app.Get(DocsPath, ui)
}

// document writes the embedded OpenAPI document.
func document(ctx *fiber.Ctx) error {
	ctx.Type("json")
	return ctx.Send(openapi.Document())
}

// ui writes a Swagger UI page backed by the embedded document route.
func ui(ctx *fiber.Ctx) error {
	ctx.Type("html")
	return ctx.SendString(swaggerHTML())
}

// swaggerHTML returns the Swagger UI HTML.
func swaggerHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>GameHub API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: "` + OpenAPIPath + `", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`
}
