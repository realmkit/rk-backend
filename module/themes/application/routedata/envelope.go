package routedata

import (
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// envelope creates the common route-data response shape.
func envelope(request Request, contract Contract) domain.RouteDataEnvelope {
	return domain.RouteDataEnvelope{
		Page:        pageData(request, contract),
		Request:     requestData(request),
		Viewer:      viewerData(request.Viewer),
		Theme:       themeData(request.Theme),
		Settings:    request.Theme.SettingsData,
		Navigation:  navigationData(request),
		Data:        routePayload(request, contract),
		Metadata:    metadataData(request, contract),
		Assets:      assetsData(request.Theme),
		Permissions: permissionsData(request.Viewer),
		Pagination:  map[string]any{},
	}
}

// pageData returns shared page data.
func pageData(request Request, contract Contract) map[string]any {
	return map[string]any{
		"route":       contract.Route,
		"title":       contract.Title,
		"description": contract.Description,
		"template":    contract.Template,
		"locale":      locale(request.Locale),
	}
}

// requestData returns safe request data.
func requestData(request Request) map[string]any {
	return map[string]any{
		"path":       request.Path,
		"params":     request.Params,
		"query":      request.Query,
		"request_id": request.RequestID,
		"now":        request.Now,
	}
}

// viewerData returns public viewer data.
func viewerData(viewer ViewerContext) map[string]any {
	return map[string]any{
		"persona":        persona(viewer.PersonaKind),
		"persona_source": personaSource(viewer.PersonaSource),
		"user_id":        viewer.UserID,
		"group_id":       viewer.GroupID,
		"preview":        viewer.IsPreview,
		"authenticated":  viewer.UserID != nil || persona(viewer.PersonaKind) != domain.PersonaAnonymous,
	}
}

// themeData returns active theme identifiers.
func themeData(theme ThemeContext) map[string]any {
	return map[string]any{
		"theme_id":         theme.ThemeID,
		"version_id":       theme.VersionID,
		"activation_id":    theme.ActivationID,
		"environment":      theme.Environment,
		"integrity_sha256": theme.IntegritySHA256,
	}
}

// navigationData returns a stable navigation shell.
func navigationData(request Request) map[string]any {
	return map[string]any{"active_route": request.Route, "items": []map[string]any{}}
}

// metadataData returns rendering metadata.
func metadataData(request Request, contract Contract) map[string]any {
	return map[string]any{
		"rich_text":       contract.RichTextFields,
		"required_params": contract.RequiredParams,
		"generated_at":    request.Now,
	}
}

// assetsData returns route asset context.
func assetsData(theme ThemeContext) map[string]any {
	return map[string]any{"theme_version_id": theme.VersionID}
}

// permissionsData returns high-level route permissions.
func permissionsData(viewer ViewerContext) map[string]any {
	return map[string]any{
		"preview":   viewer.IsPreview,
		"moderator": persona(viewer.PersonaKind) == domain.PersonaModerator,
		"signed_in": persona(viewer.PersonaKind) != domain.PersonaAnonymous,
	}
}

// locale returns a safe locale default.
func locale(value string) string {
	if value == "" {
		return "en"
	}
	return value
}

// persona returns a safe persona default.
func persona(value domain.PreviewPersonaKind) domain.PreviewPersonaKind {
	if value == "" {
		return domain.PersonaAnonymous
	}
	return value
}

// personaSource returns a safe source default.
func personaSource(value domain.PreviewPersonaSource) domain.PreviewPersonaSource {
	if value == "" {
		return domain.PersonaSourceSynthetic
	}
	return value
}
