package postgres

import (
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// themeModel maps a domain theme to persistence.
func themeModel(theme domain.Theme) ThemeModel {
	return ThemeModel{
		ID:              orm.ID{ID: theme.ID},
		Key:             string(theme.Key),
		Name:            theme.Name,
		Description:     theme.Description,
		Status:          string(theme.Status),
		CreatedByUserID: theme.CreatedBy,
		UpdatedByUserID: theme.UpdatedBy,
		Version:         theme.Version,
	}
}

// themeFromModel maps persistence to a domain theme.
func themeFromModel(model ThemeModel) domain.Theme {
	return domain.Theme{
		ID:          model.ID.ID,
		Key:         domain.Key(model.Key),
		Name:        model.Name,
		Description: model.Description,
		Status:      domain.ThemeStatus(model.Status),
		Version:     model.Version,
		CreatedBy:   model.CreatedByUserID,
		UpdatedBy:   model.UpdatedByUserID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// versionModel maps a domain version to persistence.
func versionModel(version domain.ThemeVersion) VersionModel {
	return VersionModel{
		ID:                 orm.ID{ID: version.ID},
		ThemeID:            version.ThemeID,
		Semver:             version.Semver,
		Label:              version.Label,
		Status:             string(version.Status),
		SourceKind:         string(version.SourceKind),
		SourceReference:    version.SourceReference,
		PackageStorageKey:  version.PackageStorageKey,
		PackageSizeBytes:   version.PackageSizeBytes,
		ManifestJSON:       jsonString(version.ManifestJSON),
		SettingsSchemaJSON: jsonString(version.SettingsSchemaJSON),
		SettingsDataJSON:   jsonString(version.SettingsDataJSON),
		IntegritySHA256:    string(version.IntegritySHA256),
		PublishedAt:        version.PublishedAt,
		PublishedByUserID:  version.PublishedBy,
		CreatedByUserID:    version.CreatedBy,
		UpdatedByUserID:    version.UpdatedBy,
		Version:            version.Version,
	}
}

// versionFromModel maps persistence to a domain version.
func versionFromModel(model VersionModel) domain.ThemeVersion {
	return domain.ThemeVersion{
		ID:                 model.ID.ID,
		ThemeID:            model.ThemeID,
		Semver:             model.Semver,
		Label:              model.Label,
		Status:             domain.VersionStatus(model.Status),
		SourceKind:         domain.SourceKind(model.SourceKind),
		SourceReference:    model.SourceReference,
		PackageStorageKey:  model.PackageStorageKey,
		PackageSizeBytes:   model.PackageSizeBytes,
		ManifestJSON:       []byte(model.ManifestJSON),
		SettingsSchemaJSON: []byte(model.SettingsSchemaJSON),
		SettingsDataJSON:   []byte(model.SettingsDataJSON),
		IntegritySHA256:    domain.Digest(model.IntegritySHA256),
		PublishedAt:        model.PublishedAt,
		PublishedBy:        model.PublishedByUserID,
		Version:            model.Version,
		CreatedBy:          model.CreatedByUserID,
		UpdatedBy:          model.UpdatedByUserID,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

// fileModel maps a domain file to persistence.
func fileModel(file domain.ThemeFile) FileModel {
	return FileModel{
		ID:                orm.ID{ID: file.ID},
		VersionID:         file.VersionID,
		Kind:              string(file.Kind),
		Path:              string(domain.NormalizeFilePath(file.Path)),
		ContentSHA256:     string(file.ContentSHA256),
		ContentStorageKey: file.ContentStorage,
		ContentText:       file.ContentText,
		SizeBytes:         file.SizeBytes,
	}
}

// fileFromModel maps persistence to a domain file.
func fileFromModel(model FileModel) domain.ThemeFile {
	return domain.ThemeFile{
		ID:             model.ID.ID,
		VersionID:      model.VersionID,
		Kind:           domain.FileKind(model.Kind),
		Path:           domain.FilePath(model.Path),
		ContentSHA256:  domain.Digest(model.ContentSHA256),
		ContentStorage: model.ContentStorageKey,
		ContentText:    model.ContentText,
		SizeBytes:      model.SizeBytes,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

// assetModel maps a domain asset to persistence.
func assetModel(asset domain.ThemeAsset) AssetModel {
	return AssetModel{
		ID:             orm.ID{ID: asset.ID},
		VersionID:      asset.VersionID,
		FileID:         asset.FileID,
		Path:           string(domain.NormalizeFilePath(asset.Path)),
		ContentType:    asset.ContentType,
		SizeBytes:      asset.SizeBytes,
		ContentSHA256:  string(asset.ContentSHA256),
		StorageKey:     asset.StorageKey,
		PublicURL:      asset.PublicURL,
		IntegrityValue: asset.IntegrityValue,
	}
}

// assetFromModel maps persistence to a domain asset.
func assetFromModel(model AssetModel) domain.ThemeAsset {
	return domain.ThemeAsset{
		ID:             model.ID.ID,
		VersionID:      model.VersionID,
		FileID:         model.FileID,
		Path:           domain.FilePath(model.Path),
		ContentType:    model.ContentType,
		SizeBytes:      model.SizeBytes,
		ContentSHA256:  domain.Digest(model.ContentSHA256),
		StorageKey:     model.StorageKey,
		PublicURL:      model.PublicURL,
		IntegrityValue: model.IntegrityValue,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

// issueModel maps a domain validation issue to persistence.
func issueModel(issue domain.ThemeValidationIssue) IssueModel {
	return IssueModel{
		ID:           orm.ID{ID: issue.ID},
		VersionID:    issue.VersionID,
		Severity:     string(issue.Severity),
		Code:         string(issue.Code),
		Path:         string(domain.NormalizeFilePath(issue.Path)),
		Message:      issue.Message,
		Line:         issue.Line,
		ColumnNumber: issue.Column,
		DetailsJSON:  jsonString(issue.Details),
	}
}

// issueFromModel maps persistence to a domain validation issue.
func issueFromModel(model IssueModel) domain.ThemeValidationIssue {
	return domain.ThemeValidationIssue{
		ID:        model.ID.ID,
		VersionID: model.VersionID,
		Severity:  domain.ValidationSeverity(model.Severity),
		Code:      domain.ValidationIssueCode(model.Code),
		Path:      domain.FilePath(model.Path),
		Message:   model.Message,
		Line:      model.Line,
		Column:    model.ColumnNumber,
		Details:   []byte(model.DetailsJSON),
		CreatedAt: model.CreatedAt,
	}
}

// signatureModel maps signature verification data to persistence.
func signatureModel(signature domain.ThemePackageSignature) SignatureModel {
	return SignatureModel{
		ID:                 orm.ID{ID: signature.ID},
		VersionID:          signature.VersionID,
		KeyID:              signature.KeyID,
		Algorithm:          string(signature.Algorithm),
		VerificationStatus: string(signature.VerificationStatus),
		Signature:          signature.Signature,
		SignedManifestHash: string(signature.SignedManifestHash),
		VerifiedAt:         signature.VerifiedAt,
	}
}

// signatureFromModel maps persistence to signature verification data.
func signatureFromModel(model SignatureModel) domain.ThemePackageSignature {
	return domain.ThemePackageSignature{
		ID:                 model.ID.ID,
		VersionID:          model.VersionID,
		KeyID:              model.KeyID,
		Algorithm:          domain.SignatureAlgorithm(model.Algorithm),
		VerificationStatus: domain.SignatureVerificationStatus(model.VerificationStatus),
		Signature:          model.Signature,
		SignedManifestHash: domain.Digest(model.SignedManifestHash),
		VerifiedAt:         model.VerifiedAt,
		CreatedAt:          model.CreatedAt,
	}
}

// jsonString returns an object literal for absent JSON blobs.
func jsonString(value []byte) string {
	if len(value) == 0 {
		return "{}"
	}
	return string(value)
}
