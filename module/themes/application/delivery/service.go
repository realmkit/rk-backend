package delivery

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

const (
	// immutableCacheControl is used for content-addressed version delivery.
	immutableCacheControl = "public, max-age=31536000, immutable"
	// revalidateCacheControl is used for mutable active pointers.
	revalidateCacheControl = "public, max-age=60, stale-while-revalidate=300"
	// noStoreCacheControl is used for drafts and preview tokens.
	noStoreCacheControl = "no-store"
)

// NewService creates a delivery service.
func NewService(repositories Repositories, clock Clock) Service {
	if clock == nil {
		clock = time.Now
	}
	return Service{repositories: repositories, clock: clock}
}

// ActiveActivation returns the active theme pointer for one environment.
func (service Service) ActiveActivation(
	ctx context.Context,
	environment domain.ActivationEnvironment,
) (ActivationResult, error) {
	activation, err := service.repositories.Activations.Current(ctx, environment)
	if err != nil {
		return ActivationResult{}, err
	}
	version, theme, err := service.versionTheme(ctx, activation.VersionID)
	if err != nil {
		return ActivationResult{}, err
	}
	return ActivationResult{
		Activation: activation,
		Theme:      theme,
		Version:    version,
		Cache:      activeCache(activation),
	}, nil
}

// Manifest returns the render manifest for one theme version.
func (service Service) Manifest(
	ctx context.Context,
	themeID uuid.UUID,
	versionID uuid.UUID,
) (ManifestResult, error) {
	version, theme, err := service.versionTheme(ctx, versionID)
	if err != nil {
		return ManifestResult{}, err
	}
	if theme.ID != themeID {
		return ManifestResult{}, port.ErrNotFound
	}
	files, err := service.repositories.Files.ListByVersion(ctx, version.ID)
	if err != nil {
		return ManifestResult{}, err
	}
	assets, err := service.repositories.Assets.ListByVersion(ctx, version.ID)
	if err != nil {
		return ManifestResult{}, err
	}
	return buildManifest(theme, version, files, assets), nil
}

// File returns one source file by normalized path.
func (service Service) File(ctx context.Context, versionID uuid.UUID, path domain.FilePath) (FileResult, error) {
	version, err := service.repositories.Versions.FindByID(ctx, versionID)
	if err != nil {
		return FileResult{}, err
	}
	file, err := service.repositories.Files.FindByPath(ctx, version.ID, path)
	if err != nil {
		return FileResult{}, err
	}
	return FileResult{File: file, Cache: fileCache(version, file)}, nil
}

// Asset returns immutable asset metadata by normalized path.
func (service Service) Asset(ctx context.Context, versionID uuid.UUID, path domain.FilePath) (AssetResult, error) {
	assets, err := service.repositories.Assets.ListByVersion(ctx, versionID)
	if err != nil {
		return AssetResult{}, err
	}
	for _, asset := range assets {
		if domain.NormalizeFilePath(asset.Path) == domain.NormalizeFilePath(path) {
			return AssetResult{Asset: asset, Cache: assetCache(asset)}, nil
		}
	}
	return AssetResult{}, port.ErrNotFound
}

// ValidationReport returns static validation diagnostics.
func (service Service) ValidationReport(ctx context.Context, versionID uuid.UUID) (ValidationReport, error) {
	version, err := service.repositories.Versions.FindByID(ctx, versionID)
	if err != nil {
		return ValidationReport{}, err
	}
	issues, err := service.repositories.Issues.ListByVersion(ctx, version.ID)
	if err != nil {
		return ValidationReport{}, err
	}
	return ValidationReport{
		Version:       version,
		Issues:        issues,
		RouteCoverage: manifestObject(version.ManifestJSON, "route_coverage"),
		Cache:         versionCache(version),
	}, nil
}

// versionTheme returns a version with its theme family.
func (service Service) versionTheme(ctx context.Context, versionID uuid.UUID) (domain.ThemeVersion, domain.Theme, error) {
	version, err := service.repositories.Versions.FindByID(ctx, versionID)
	if err != nil {
		return domain.ThemeVersion{}, domain.Theme{}, err
	}
	theme, err := service.repositories.Themes.FindByID(ctx, version.ThemeID)
	if err != nil {
		return domain.ThemeVersion{}, domain.Theme{}, err
	}
	return version, theme, nil
}
