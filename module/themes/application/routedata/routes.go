package routedata

import (
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Contracts returns every first-version route-data contract.
func Contracts() []Contract {
	return []Contract{
		route(domain.RouteHome, "Home"),
		route(domain.RouteForumsIndex, "Forums"),
		richRoute(domain.RouteForumsCategory, "Forum category", []string{"category_slug"}, forumDescription()),
		richRoute(domain.RouteForumsShow, "Forum", []string{"forum_slug"}, forumDescription()),
		richRoute(domain.RouteThreadsShow, "Thread", []string{"thread_slug"}, forumPost()),
		route(domain.RouteThreadsNew, "New thread"),
		route(domain.RouteTicketsIndex, "Tickets"),
		richRoute(domain.RouteTicketsNew, "New ticket", nil, ticketText()),
		richRoute(domain.RouteTicketsShow, "Ticket", []string{"ticket_id"}, ticketText()),
		richRoute(domain.RoutePunishmentsIndex, "Punishments", nil, punishmentText()),
		richRoute(domain.RoutePunishmentsShow, "Punishment", []string{"punishment_id"}, punishmentText()),
		richRoute(domain.RouteUsersShow, "User profile", []string{"user_id_or_slug"}, signatureText()),
		route(domain.RouteSearch, "Search"),
		richRoute(domain.RouteStaticPage, "Static page", []string{"page_slug"}, staticPage()),
		route(domain.RouteNotFound, "Not found"),
		route(domain.RouteError, "Error"),
		route(domain.RouteMaintenance, "Maintenance"),
		route(domain.RouteLogin, "Login"),
		route(domain.RouteRegister, "Register"),
		route(domain.RouteForgotPassword, "Forgot password"),
		route(domain.RouteResetPassword, "Reset password"),
		route(domain.RouteVerifyEmail, "Verify email"),
		route(domain.RouteAccountRecovery, "Account recovery"),
	}
}

// ContractFor returns one route-data contract.
func ContractFor(kind domain.RouteKind) (Contract, error) {
	for _, contract := range Contracts() {
		if contract.Route == kind {
			return contract, nil
		}
	}
	return Contract{}, port.ErrNotFound
}

// route returns a simple route contract.
func route(kind domain.RouteKind, title string) Contract {
	return richRoute(kind, title, nil, nil)
}

// richRoute returns a route contract with rich-text fields.
func richRoute(
	kind domain.RouteKind,
	title string,
	required []string,
	richText map[string]domain.RichTextProfile,
) Contract {
	return Contract{
		Route:          kind,
		Template:       domain.FilePath("templates/" + string(kind) + ".liquid"),
		Title:          title,
		Description:    title + " route data.",
		RequiredParams: required,
		RichTextFields: richText,
	}
}

// forumDescription returns forum description rich text fields.
func forumDescription() map[string]domain.RichTextProfile {
	return map[string]domain.RichTextProfile{"description_html": domain.ProfileForumDescription}
}

// forumPost returns forum post rich text fields.
func forumPost() map[string]domain.RichTextProfile {
	return map[string]domain.RichTextProfile{"posts[].content_html": domain.ProfileForumPost}
}

// ticketText returns ticket rich text fields.
func ticketText() map[string]domain.RichTextProfile {
	return map[string]domain.RichTextProfile{"messages[].content_html": domain.ProfileTicketText}
}

// punishmentText returns punishment rich text fields.
func punishmentText() map[string]domain.RichTextProfile {
	return map[string]domain.RichTextProfile{"description_html": domain.ProfilePunishmentText}
}

// signatureText returns user signature rich text fields.
func signatureText() map[string]domain.RichTextProfile {
	return map[string]domain.RichTextProfile{"signature_html": domain.ProfileSignature}
}

// staticPage returns static page rich text fields.
func staticPage() map[string]domain.RichTextProfile {
	return map[string]domain.RichTextProfile{"body_html": domain.ProfileStaticPage}
}
