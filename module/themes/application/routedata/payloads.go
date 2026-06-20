package routedata

import "github.com/realmkit/rk-backend/module/themes/domain"

// routePayload returns route-family-specific data shells.
func routePayload(request Request, contract Contract) map[string]any {
	base := map[string]any{"kind": contract.Route, "params": request.Params}
	switch contract.Route {
	case domain.RouteForumsIndex:
		base["categories"] = []map[string]any{}
	case domain.RouteForumsCategory:
		base["category"] = objectFromParam("slug", request.Params["category_slug"])
		base["forums"] = []map[string]any{}
	case domain.RouteForumsShow:
		base["forum"] = objectFromParam("slug", request.Params["forum_slug"])
		base["threads"] = []map[string]any{}
	case domain.RouteThreadsShow:
		base["thread"] = objectFromParam("slug", request.Params["thread_slug"])
		base["posts"] = []map[string]any{}
	case domain.RouteThreadsNew:
		base["form"] = map[string]any{"kind": "thread"}
	case domain.RouteTicketsIndex:
		base["tickets"] = []map[string]any{}
	case domain.RouteTicketsNew:
		base["form"] = map[string]any{"kind": "ticket"}
	case domain.RouteTicketsShow:
		base["ticket"] = objectFromParam("id", request.Params["ticket_id"])
	case domain.RoutePunishmentsIndex:
		base["punishments"] = []map[string]any{}
	case domain.RoutePunishmentsShow:
		base["punishment"] = objectFromParam("id", request.Params["punishment_id"])
	case domain.RouteUsersShow:
		base["user"] = objectFromParam("id_or_slug", request.Params["user_id_or_slug"])
	case domain.RouteSearch:
		base["results"] = []map[string]any{}
		base["query"] = request.Query["q"]
	case domain.RouteStaticPage:
		base["page"] = objectFromParam("slug", request.Params["page_slug"])
	case domain.RouteHome:
		base["sections"] = []map[string]any{}
	default:
		base["state"] = string(contract.Route)
	}
	return base
}

// objectFromParam returns an object keyed by one route parameter.
func objectFromParam(key string, value string) map[string]any {
	return map[string]any{key: value}
}
