package delivery

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by theme delivery.
type Repositories struct {
	Themes        port.ThemeRepository           // Themes stores the themes value.
	Versions      port.VersionRepository         // Versions stores the versions value.
	Files         port.FileRepository            // Files stores the files value.
	Assets        port.AssetRepository           // Assets stores the assets value.
	Activations   port.ActivationRepository      // Activations stores the activations value.
	Issues        port.ValidationIssueRepository // Issues stores the issues value.
	PreviewTokens port.PreviewTokenRepository    // PreviewTokens stores the preview tokens value.
}

// CacheMetadata describes HTTP cache headers for a delivery response.
type CacheMetadata struct {
	ETag          string        // ETag stores the e tag value.
	CacheControl  string        // CacheControl stores the cache control value.
	LastModified  time.Time     // LastModified stores the last modified value.
	ContentSHA256 domain.Digest // ContentSHA256 stores the content s h a256 value.
}

// ActivationResult contains the active theme pointer and related records.
type ActivationResult struct {
	Activation domain.ThemeActivation // Activation stores the activation value.
	Theme      domain.Theme           // Theme stores the theme value.
	Version    domain.ThemeVersion    // Version stores the version value.
	Cache      CacheMetadata          // Cache stores the cache value.
}

// ManifestResult contains a renderable theme manifest for Next.js.
type ManifestResult struct {
	Theme           domain.Theme                // Theme stores the theme value.
	Version         domain.ThemeVersion         // Version stores the version value.
	Files           []domain.ThemeFile          // Files stores the files value.
	Assets          map[string]map[string]any   // Assets stores the assets value.
	Layouts         map[string]domain.ThemeFile // Layouts stores the layouts value.
	Templates       map[string]domain.ThemeFile // Templates stores the templates value.
	Sections        map[string]domain.ThemeFile // Sections stores the sections value.
	Snippets        map[string]domain.ThemeFile // Snippets stores the snippets value.
	Locales         map[string]domain.ThemeFile // Locales stores the locales value.
	SettingsSchema  map[string]any              // SettingsSchema stores the settings schema value.
	SettingsData    map[string]any              // SettingsData stores the settings data value.
	DependencyGraph map[string]any              // DependencyGraph stores the dependency graph value.
	RouteCoverage   map[string]any              // RouteCoverage stores the route coverage value.
	CSP             map[string]any              // CSP stores the c s p value.
	Cache           CacheMetadata               // Cache stores the cache value.
}

// FileResult contains one theme source file and cache metadata.
type FileResult struct {
	File  domain.ThemeFile // File stores the file value.
	Cache CacheMetadata    // Cache stores the cache value.
}

// AssetResult contains one immutable theme asset and cache metadata.
type AssetResult struct {
	Asset domain.ThemeAsset // Asset stores the asset value.
	Cache CacheMetadata     // Cache stores the cache value.
}

// ValidationReport contains version validation diagnostics.
type ValidationReport struct {
	Version       domain.ThemeVersion           // Version stores the version value.
	Issues        []domain.ThemeValidationIssue // Issues stores the issues value.
	RouteCoverage map[string]any                // RouteCoverage stores the route coverage value.
	Cache         CacheMetadata                 // Cache stores the cache value.
}

// PreviewTokenResult returns the raw token exactly once.
type PreviewTokenResult struct {
	Token     string                   // Token stores the token value.
	Preview   domain.ThemePreviewToken // Preview stores the preview value.
	ExpiresAt time.Time                // ExpiresAt stores the expires at value.
}

// CreatePreviewTokenCommand requests a scoped preview token.
type CreatePreviewTokenCommand struct {
	VersionID     uuid.UUID                   // VersionID stores the version i d value.
	PersonaKind   domain.PreviewPersonaKind   // PersonaKind stores the persona kind value.
	PersonaSource domain.PreviewPersonaSource // PersonaSource stores the persona source value.
	PersonaUserID *uuid.UUID                  // PersonaUserID stores the persona user i d value.
	TTL           time.Duration               // TTL stores the t t l value.
	ActorUserID   *uuid.UUID                  // ActorUserID stores the actor user i d value.
}

// ValidatePreviewTokenCommand validates a raw preview token.
type ValidatePreviewTokenCommand struct {
	Token string // Token stores the token value.
}

// Clock returns the current time.
type Clock func() time.Time

// Service owns cacheable theme delivery and preview token workflows.
type Service struct {
	repositories Repositories // repositories stores the repositories value.
	clock        Clock        // clock stores the clock value.
}
