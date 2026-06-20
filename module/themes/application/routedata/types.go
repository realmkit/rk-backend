package routedata

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// ThemeContext describes the selected theme version for route rendering.
type ThemeContext struct {
	ThemeID         uuid.UUID
	VersionID       uuid.UUID
	ActivationID    uuid.UUID
	Environment     domain.ActivationEnvironment
	SettingsData    map[string]any
	IntegritySHA256 domain.Digest
}

// ViewerContext describes the viewer lens used for route data.
type ViewerContext struct {
	PersonaKind   domain.PreviewPersonaKind
	PersonaSource domain.PreviewPersonaSource
	UserID        *uuid.UUID
	GroupID       *uuid.UUID
	IsPreview     bool
}

// Request asks for one public route-data envelope.
type Request struct {
	Route     domain.RouteKind
	Locale    string
	Path      string
	Params    map[string]string
	Query     map[string]string
	Theme     ThemeContext
	Viewer    ViewerContext
	Now       time.Time
	RequestID string
}

// Contract describes one route-data contract.
type Contract struct {
	Route          domain.RouteKind
	Template       domain.FilePath
	Title          string
	Description    string
	RequiredParams []string
	RichTextFields map[string]domain.RichTextProfile
}

// VisibilityChecker verifies route visibility before data is returned.
type VisibilityChecker interface {
	CanViewRoute(context.Context, Request, Contract) error
}

// Service owns public route-data envelope construction.
type Service struct {
	visibility VisibilityChecker
	clock      Clock
}

// Clock returns the current time.
type Clock func() time.Time
