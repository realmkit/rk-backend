package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/gorm"
)

// FileRepository stores version files in PostgreSQL.
type FileRepository struct {
	store orm.Store
}

// NewFileRepository creates a theme file repository.
func NewFileRepository(store orm.Store) FileRepository {
	return FileRepository{store: store}
}

// ReplaceVersionFiles replaces active files for a version.
func (repository FileRepository) ReplaceVersionFiles(
	ctx context.Context,
	versionID uuid.UUID,
	files []domain.ThemeFile,
) error {
	if err := ensureVersionEditable(ctx, repository.store, versionID); err != nil {
		return err
	}
	return repository.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("version_id = ?", versionID).Delete(&FileModel{}).Error; err != nil {
			return err
		}
		for _, file := range files {
			file.VersionID = versionID
			model := fileModel(file)
			if err := tx.Create(&model).Error; err != nil {
				return port.ErrConflict
			}
		}
		return nil
	})
}

// ListByVersion returns active files for a version.
func (repository FileRepository) ListByVersion(ctx context.Context, versionID uuid.UUID) ([]domain.ThemeFile, error) {
	var models []FileModel
	err := repository.store.DB(ctx).
		Where("version_id = ?", versionID).
		Order("path ASC").
		Find(&models).
		Error
	if err != nil {
		return nil, err
	}
	files := make([]domain.ThemeFile, 0, len(models))
	for _, model := range models {
		files = append(files, fileFromModel(model))
	}
	return files, nil
}

// FindByPath returns one active file by normalized path.
func (repository FileRepository) FindByPath(
	ctx context.Context,
	versionID uuid.UUID,
	filePath domain.FilePath,
) (domain.ThemeFile, error) {
	var model FileModel
	err := repository.store.DB(ctx).
		First(&model, "version_id = ? AND path = ?", versionID, domain.NormalizeFilePath(filePath)).
		Error
	if err != nil {
		return domain.ThemeFile{}, mapError(err)
	}
	return fileFromModel(model), nil
}

// AssetRepository stores version assets in PostgreSQL.
type AssetRepository struct {
	store orm.Store
}

// NewAssetRepository creates a theme asset repository.
func NewAssetRepository(store orm.Store) AssetRepository {
	return AssetRepository{store: store}
}

// ReplaceVersionAssets replaces active assets for a version.
func (repository AssetRepository) ReplaceVersionAssets(
	ctx context.Context,
	versionID uuid.UUID,
	assets []domain.ThemeAsset,
) error {
	if err := ensureVersionEditable(ctx, repository.store, versionID); err != nil {
		return err
	}
	return repository.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("version_id = ?", versionID).Delete(&AssetModel{}).Error; err != nil {
			return err
		}
		for _, asset := range assets {
			asset.VersionID = versionID
			model := assetModel(asset)
			if err := tx.Create(&model).Error; err != nil {
				return port.ErrConflict
			}
		}
		return nil
	})
}

// ListByVersion returns active assets for a version.
func (repository AssetRepository) ListByVersion(ctx context.Context, versionID uuid.UUID) ([]domain.ThemeAsset, error) {
	var models []AssetModel
	err := repository.store.DB(ctx).
		Where("version_id = ?", versionID).
		Order("path ASC").
		Find(&models).
		Error
	if err != nil {
		return nil, err
	}
	assets := make([]domain.ThemeAsset, 0, len(models))
	for _, model := range models {
		assets = append(assets, assetFromModel(model))
	}
	return assets, nil
}

// ValidationIssueRepository stores validation diagnostics in PostgreSQL.
type ValidationIssueRepository struct {
	store orm.Store
}

// NewValidationIssueRepository creates a validation issue repository.
func NewValidationIssueRepository(store orm.Store) ValidationIssueRepository {
	return ValidationIssueRepository{store: store}
}

// ReplaceVersionIssues replaces active diagnostics for a version.
func (repository ValidationIssueRepository) ReplaceVersionIssues(
	ctx context.Context,
	versionID uuid.UUID,
	issues []domain.ThemeValidationIssue,
) error {
	if err := ensureVersionEditable(ctx, repository.store, versionID); err != nil {
		return err
	}
	return repository.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("version_id = ?", versionID).Delete(&IssueModel{}).Error; err != nil {
			return err
		}
		for _, issue := range issues {
			issue.VersionID = versionID
			model := issueModel(issue)
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListByVersion returns active diagnostics for a version.
func (repository ValidationIssueRepository) ListByVersion(
	ctx context.Context,
	versionID uuid.UUID,
) ([]domain.ThemeValidationIssue, error) {
	var models []IssueModel
	err := repository.store.DB(ctx).
		Where("version_id = ?", versionID).
		Order("severity ASC, code ASC, path ASC, line ASC").
		Find(&models).
		Error
	if err != nil {
		return nil, err
	}
	issues := make([]domain.ThemeValidationIssue, 0, len(models))
	for _, model := range models {
		issues = append(issues, issueFromModel(model))
	}
	return issues, nil
}

// SignatureRepository stores signature verification data in PostgreSQL.
type SignatureRepository struct {
	store orm.Store
}

// NewSignatureRepository creates a package signature repository.
func NewSignatureRepository(store orm.Store) SignatureRepository {
	return SignatureRepository{store: store}
}

// ReplaceVersionSignature replaces signature data for a version.
func (repository SignatureRepository) ReplaceVersionSignature(
	ctx context.Context,
	versionID uuid.UUID,
	signature domain.ThemePackageSignature,
) error {
	if err := ensureVersionEditable(ctx, repository.store, versionID); err != nil {
		return err
	}
	return repository.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("version_id = ?", versionID).Delete(&SignatureModel{}).Error; err != nil {
			return err
		}
		signature.VersionID = versionID
		model := signatureModel(signature)
		if err := tx.Create(&model).Error; err != nil {
			return port.ErrConflict
		}
		return nil
	})
}

// FindByVersion returns package signature data for a version.
func (repository SignatureRepository) FindByVersion(
	ctx context.Context,
	versionID uuid.UUID,
) (domain.ThemePackageSignature, error) {
	var model SignatureModel
	if err := repository.store.DB(ctx).First(&model, "version_id = ?", versionID).Error; err != nil {
		return domain.ThemePackageSignature{}, mapError(err)
	}
	return signatureFromModel(model), nil
}

// ensureVersionEditable verifies a version can receive content writes.
func ensureVersionEditable(ctx context.Context, store orm.Store, versionID uuid.UUID) error {
	var model VersionModel
	if err := store.DB(ctx).First(&model, "id = ?", versionID).Error; err != nil {
		return mapError(err)
	}
	return versionFromModel(model).EnsureEditable()
}

// Ensure content repositories implement their ports.
var (
	_ port.FileRepository            = FileRepository{}
	_ port.AssetRepository           = AssetRepository{}
	_ port.ValidationIssueRepository = ValidationIssueRepository{}
	_ port.SignatureRepository       = SignatureRepository{}
)
