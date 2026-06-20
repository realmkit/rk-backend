package delivery

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by theme delivery.
type Repositories struct {
	Themes        port.ThemeRepository
	Versions      port.VersionRepository
	Files         port.FileRepository
	Assets        port.AssetRepository
	Activations   port.ActivationRepository
	Issues        port.ValidationIssueRepository
	PreviewTokens port.PreviewTokenRepository
}

// CacheMetadata describes HTTP cache headers for a delivery response.
type CacheMetadata struct {
	ETag          string
	CacheControl  string
	LastModified  time.Time
	ContentSHA256 domain.Digest
}

// ActivationResult contains the active theme pointer and related records.
type ActivationResult struct {
	Activation domain.ThemeActivation
	Theme      domain.Theme
	Version    domain.ThemeVersion
	Cache      CacheMetadata
}

// ManifestResult contains a renderable theme manifest for Next.js.
type ManifestResult struct {
	Theme           domain.Theme
	Version         domain.ThemeVersion
	Files           []domain.ThemeFile
	Assets          map[string]map[string]any
	Layouts         map[string]domain.ThemeFile
	Templates       map[string]domain.ThemeFile
	Sections        map[string]domain.ThemeFile
	Snippets        map[string]domain.ThemeFile
	Locales         map[string]domain.ThemeFile
	SettingsSchema  map[string]any
	SettingsData    map[string]any
	DependencyGraph map[string]any
	RouteCoverage   map[string]any
	CSP             map[string]any
	Cache           CacheMetadata
}

// FileResult contains one theme source file and cache metadata.
type FileResult struct {
	File  domain.ThemeFile
	Cache CacheMetadata
}

// AssetResult contains one immutable theme asset and cache metadata.
type AssetResult struct {
	Asset domain.ThemeAsset
	Cache CacheMetadata
}

// ValidationReport contains version validation diagnostics.
type ValidationReport struct {
	Version       domain.ThemeVersion
	Issues        []domain.ThemeValidationIssue
	RouteCoverage map[string]any
	Cache         CacheMetadata
}

// PreviewTokenResult returns the raw token exactly once.
type PreviewTokenResult struct {
	Token     string
	Preview   domain.ThemePreviewToken
	ExpiresAt time.Time
}

// CreatePreviewTokenCommand requests a scoped preview token.
type CreatePreviewTokenCommand struct {
	VersionID     uuid.UUID
	PersonaKind   domain.PreviewPersonaKind
	PersonaSource domain.PreviewPersonaSource
	PersonaUserID *uuid.UUID
	TTL           time.Duration
	ActorUserID   *uuid.UUID
}

// ValidatePreviewTokenCommand validates a raw preview token.
type ValidatePreviewTokenCommand struct {
	Token string
}

// Clock returns the current time.
type Clock func() time.Time

// Service owns cacheable theme delivery and preview token workflows.
type Service struct {
	repositories Repositories
	clock        Clock
}
