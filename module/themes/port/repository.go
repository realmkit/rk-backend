package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// ThemeFilter filters theme families.
type ThemeFilter struct {
	Status domain.ThemeStatus // Status stores the status value.
}

// ThemeRepository stores theme families.
type ThemeRepository interface {
	Create(context.Context, domain.Theme) (domain.Theme, error)
	Update(context.Context, domain.Theme, uint64) (domain.Theme, error)
	Archive(context.Context, uuid.UUID, uint64) error
	FindByID(context.Context, uuid.UUID) (domain.Theme, error)
	List(context.Context, ThemeFilter) ([]domain.Theme, error)
}

// VersionRepository stores theme versions.
type VersionRepository interface {
	Create(context.Context, domain.ThemeVersion) (domain.ThemeVersion, error)
	Update(context.Context, domain.ThemeVersion, uint64) (domain.ThemeVersion, error)
	Archive(context.Context, uuid.UUID, uint64) error
	FindByID(context.Context, uuid.UUID) (domain.ThemeVersion, error)
	FindBySourceReference(context.Context, uuid.UUID, string) (domain.ThemeVersion, error)
	ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeVersion, error)
}

// FileRepository stores version files.
type FileRepository interface {
	ReplaceVersionFiles(context.Context, uuid.UUID, []domain.ThemeFile) error
	ListByVersion(context.Context, uuid.UUID) ([]domain.ThemeFile, error)
	FindByPath(context.Context, uuid.UUID, domain.FilePath) (domain.ThemeFile, error)
}

// AssetRepository stores derived version assets.
type AssetRepository interface {
	ReplaceVersionAssets(context.Context, uuid.UUID, []domain.ThemeAsset) error
	ListByVersion(context.Context, uuid.UUID) ([]domain.ThemeAsset, error)
}

// ActivationRepository stores active version pointers.
type ActivationRepository interface {
	Activate(context.Context, domain.ThemeActivation) (domain.ThemeActivation, error)
	Current(context.Context, domain.ActivationEnvironment) (domain.ThemeActivation, error)
	FindByID(context.Context, uuid.UUID) (domain.ThemeActivation, error)
	ListByTheme(context.Context, uuid.UUID) ([]domain.ThemeActivation, error)
}

// ValidationIssueRepository stores version diagnostics.
type ValidationIssueRepository interface {
	ReplaceVersionIssues(context.Context, uuid.UUID, []domain.ThemeValidationIssue) error
	ListByVersion(context.Context, uuid.UUID) ([]domain.ThemeValidationIssue, error)
}

// SignatureRepository stores package signature verification data.
type SignatureRepository interface {
	ReplaceVersionSignature(context.Context, uuid.UUID, domain.ThemePackageSignature) error
	FindByVersion(context.Context, uuid.UUID) (domain.ThemePackageSignature, error)
}

// SigningKeyRepository stores trusted package signing keys.
type SigningKeyRepository interface {
	Upsert(context.Context, domain.ThemeSigningKey) (domain.ThemeSigningKey, error)
	FindByKeyID(context.Context, string) (domain.ThemeSigningKey, error)
	List(context.Context) ([]domain.ThemeSigningKey, error)
}

// PreviewTokenRepository stores short-lived preview tokens.
type PreviewTokenRepository interface {
	Create(context.Context, domain.ThemePreviewToken) (domain.ThemePreviewToken, error)
	FindByTokenHash(context.Context, string) (domain.ThemePreviewToken, error)
	Revoke(context.Context, uuid.UUID) error
}
