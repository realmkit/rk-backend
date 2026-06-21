package routedata

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// ThemeContext describes the selected theme version for route rendering.
type ThemeContext struct {
	ThemeID         uuid.UUID                    // ThemeID stores the theme i d value.
	VersionID       uuid.UUID                    // VersionID stores the version i d value.
	ActivationID    uuid.UUID                    // ActivationID stores the activation i d value.
	Environment     domain.ActivationEnvironment // Environment stores the environment value.
	SettingsData    map[string]any               // SettingsData stores the settings data value.
	IntegritySHA256 domain.Digest                // IntegritySHA256 stores the integrity s h a256 value.
}

// ViewerContext describes the viewer lens used for route data.
type ViewerContext struct {
	PersonaKind   domain.PreviewPersonaKind   // PersonaKind stores the persona kind value.
	PersonaSource domain.PreviewPersonaSource // PersonaSource stores the persona source value.
	UserID        *uuid.UUID                  // UserID stores the user i d value.
	GroupID       *uuid.UUID                  // GroupID stores the group i d value.
	IsPreview     bool                        // IsPreview stores the is preview value.
}

// Request asks for one public route-data envelope.
type Request struct {
	Route     domain.RouteKind  // Route stores the route value.
	Locale    string            // Locale stores the locale value.
	Path      string            // Path stores the path value.
	Params    map[string]string // Params stores the params value.
	Query     map[string]string // Query stores the query value.
	Theme     ThemeContext      // Theme stores the theme value.
	Viewer    ViewerContext     // Viewer stores the viewer value.
	Now       time.Time         // Now stores the now value.
	RequestID string            // RequestID stores the request i d value.
}

// Contract describes one route-data contract.
type Contract struct {
	Route          domain.RouteKind                  // Route stores the route value.
	Template       domain.FilePath                   // Template stores the template value.
	Title          string                            // Title stores the title value.
	Description    string                            // Description stores the description value.
	RequiredParams []string                          // RequiredParams stores the required params value.
	RichTextFields map[string]domain.RichTextProfile // RichTextFields stores the rich text fields value.
}

// VisibilityChecker verifies route visibility before data is returned.
type VisibilityChecker interface {
	CanViewRoute(context.Context, Request, Contract) error
}

// Service owns public route-data envelope construction.
type Service struct {
	visibility VisibilityChecker // visibility stores the visibility value.
	clock      Clock             // clock stores the clock value.
}

// Clock returns the current time.
type Clock func() time.Time
