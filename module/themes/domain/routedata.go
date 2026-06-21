package domain

const (
	// RouteHome renders the public home page.
	RouteHome RouteKind = "home"
	// RouteForumsIndex renders the forum index.
	RouteForumsIndex RouteKind = "forums.index"
	// RouteForumsCategory renders one forum category.
	RouteForumsCategory RouteKind = "forums.category"
	// RouteForumsShow renders one forum.
	RouteForumsShow RouteKind = "forums.show"
	// RouteThreadsShow renders one thread.
	RouteThreadsShow RouteKind = "threads.show"
	// RouteThreadsNew renders thread creation.
	RouteThreadsNew RouteKind = "threads.new"
	// RouteTicketsIndex renders public ticket listing.
	RouteTicketsIndex RouteKind = "tickets.index"
	// RouteTicketsNew renders ticket creation.
	RouteTicketsNew RouteKind = "tickets.new"
	// RouteTicketsShow renders one ticket.
	RouteTicketsShow RouteKind = "tickets.show"
	// RoutePunishmentsIndex renders a public punishment listing.
	RoutePunishmentsIndex RouteKind = "punishments.index"
	// RoutePunishmentsShow renders one punishment.
	RoutePunishmentsShow RouteKind = "punishments.show"
	// RouteUsersShow renders one public profile.
	RouteUsersShow RouteKind = "users.show"
	// RouteSearch renders a public search.
	RouteSearch RouteKind = "search"
	// RouteStaticPage renders a static page.
	RouteStaticPage RouteKind = "static.page"
	// RouteNotFound renders a not-found page.
	RouteNotFound RouteKind = "not_found"
	// RouteError renders a public error page.
	RouteError RouteKind = "error"
	// RouteMaintenance renders maintenance mode.
	RouteMaintenance RouteKind = "maintenance"
	// RouteLogin renders login.
	RouteLogin RouteKind = "auth.login"
	// RouteRegister renders registration.
	RouteRegister RouteKind = "auth.register"
	// RouteForgotPassword renders a password reset request.
	RouteForgotPassword RouteKind = "auth.forgot_password"
	// RouteResetPassword renders password reset completion.
	RouteResetPassword RouteKind = "auth.reset_password"
	// RouteVerifyEmail renders email verification.
	RouteVerifyEmail RouteKind = "auth.verify_email"
	// RouteAccountRecovery renders account recovery.
	RouteAccountRecovery RouteKind = "auth.account_recovery"
)

const (
	// ProfileForumPost sanitizes forum post rich text.
	ProfileForumPost RichTextProfile = "forum_post"
	// ProfileForumDescription sanitizes forum descriptions.
	ProfileForumDescription RichTextProfile = "forum_description"
	// ProfileStaticPage sanitizes static page rich text.
	ProfileStaticPage RichTextProfile = "static_page"
	// ProfileTicketText sanitizes ticket-rich text.
	ProfileTicketText RichTextProfile = "ticket_text"
	// ProfilePunishmentText sanitizes punishment rich text.
	ProfilePunishmentText RichTextProfile = "punishment_text"
	// ProfileSignature sanitizes user signatures.
	ProfileSignature RichTextProfile = "signature"
)

const (
	// PersonaAnonymous previews as a public visitor.
	PersonaAnonymous PreviewPersonaKind = "anonymous"
	// PersonaSignedIn previews as an authenticated user.
	PersonaSignedIn PreviewPersonaKind = "signed_in"
	// PersonaGroup previews as a selected group membership.
	PersonaGroup PreviewPersonaKind = "group"
	// PersonaModerator previews as staff moderation access.
	PersonaModerator PreviewPersonaKind = "moderator"
	// PersonaUser previews as a real selected user.
	PersonaUser PreviewPersonaKind = "user"
)

const (
	// PersonaSourceSynthetic indicates simulated preview data.
	PersonaSourceSynthetic PreviewPersonaSource = "synthetic"
	// PersonaSourceReal indicates data resolved from a real user.
	PersonaSourceReal PreviewPersonaSource = "real"
)

// RouteDataEnvelope describes the common route-data response shape.
type RouteDataEnvelope struct {
	Page        map[string]any   `json:"page"`                 // Page stores the page value.
	Request     map[string]any   `json:"request"`              // Request stores the request value.
	Viewer      map[string]any   `json:"viewer"`               // Viewer stores the viewer value.
	Theme       map[string]any   `json:"theme"`                // Theme stores the theme value.
	Settings    map[string]any   `json:"settings"`             // Settings stores the settings value.
	Navigation  map[string]any   `json:"navigation"`           // Navigation stores the navigation value.
	Data        map[string]any   `json:"data"`                 // Data stores the data value.
	Metadata    map[string]any   `json:"metadata"`             // Metadata stores the metadata value.
	Assets      map[string]any   `json:"assets"`               // Assets stores the assets value.
	Permissions map[string]any   `json:"permissions"`          // Permissions stores the permissions value.
	Pagination  map[string]any   `json:"pagination,omitempty"` // Pagination stores the pagination value.
	Flash       []map[string]any `json:"flash,omitempty"`      // Flash stores the flash value.
}

// RouteKinds returns all first-version route-data contracts.
func RouteKinds() []RouteKind {
	return []RouteKind{
		RouteHome,
		RouteForumsIndex,
		RouteForumsCategory,
		RouteForumsShow,
		RouteThreadsShow,
		RouteThreadsNew,
		RouteTicketsIndex,
		RouteTicketsNew,
		RouteTicketsShow,
		RoutePunishmentsIndex,
		RoutePunishmentsShow,
		RouteUsersShow,
		RouteSearch,
		RouteStaticPage,
		RouteNotFound,
		RouteError,
		RouteMaintenance,
		RouteLogin,
		RouteRegister,
		RouteForgotPassword,
		RouteResetPassword,
		RouteVerifyEmail,
		RouteAccountRecovery,
	}
}
