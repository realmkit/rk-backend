package importing

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/application/signing"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// NewService creates a theme package import service.
func NewService(
	repositories Repositories,
	store storage.Store,
	cfg Config,
	verifier signatureVerifier,
) Service {
	return Service{
		repositories: repositories,
		store:        store,
		cfg:          cfg.Defaults(),
		verifier:     verifier,
	}
}

// Import imports one uploaded theme package as a version.
func (service Service) Import(ctx context.Context, command Command) (Result, error) {
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return Result{}, fmt.Errorf("theme import idempotency key is required")
	}
	sourceReference := sourceReference(command.IdempotencyKey)
	existing, err := service.repositories.Versions.FindBySourceReference(ctx, command.ThemeID, sourceReference)
	if err == nil {
		return Result{Version: existing, Reused: true}, nil
	}
	if err != nil && err != port.ErrNotFound {
		return Result{}, err
	}
	files, issues, err := extractPackage(ctx, command.Package, command.PackageSizeBytes, service.cfg)
	if err != nil {
		return Result{}, err
	}
	manifest, manifestIssues := decodeManifest(files)
	issues = append(issues, manifestIssues...)
	signatureResult := service.verifySignature(ctx, files, manifest.raw)
	issues = append(issues, signatureResult.Issues...)
	version, err := service.createVersion(ctx, command, sourceReference, manifest)
	if err != nil {
		return Result{}, err
	}
	if !hasError(issues) {
		if err := service.persistFilesAndAssets(ctx, version.ID, files); err != nil {
			return Result{}, err
		}
	}
	if err := service.persistDiagnostics(ctx, version.ID, issues, signatureResult.Signature); err != nil {
		return Result{}, err
	}
	version, err = service.finalizeVersion(ctx, version, files, issues)
	if err != nil {
		return Result{}, err
	}
	return Result{Version: version, Signature: signatureResult.Signature, Issues: issues, Imported: true}, nil
}

// createVersion stores the initial imported version row.
func (service Service) createVersion(
	ctx context.Context,
	command Command,
	sourceReference string,
	manifest manifestPayload,
) (domain.ThemeVersion, error) {
	return service.repositories.Versions.Create(ctx, domain.ThemeVersion{
		ID:                 uuid.New(),
		ThemeID:            command.ThemeID,
		Semver:             coalesce(command.Semver, manifest.document.Version),
		Label:              coalesce(command.Label, manifest.document.Name, "Imported theme package"),
		Status:             domain.VersionStatusValidating,
		SourceKind:         domain.SourceUpload,
		SourceReference:    sourceReference,
		PackageSizeBytes:   command.PackageSizeBytes,
		ManifestJSON:       manifest.raw,
		SettingsSchemaJSON: findFileBytes(manifest.files, "config/settings_schema.json"),
		SettingsDataJSON:   []byte(`{}`),
		CreatedBy:          command.ActorUserID,
		UpdatedBy:          command.ActorUserID,
		Version:            1,
	})
}

// persistFilesAndAssets stores immutable files and derived assets.
func (service Service) persistFilesAndAssets(
	ctx context.Context,
	versionID uuid.UUID,
	files []packageFile,
) error {
	domainFiles, assets, err := service.domainFiles(ctx, versionID, files)
	if err != nil {
		return err
	}
	if err := service.repositories.Files.ReplaceVersionFiles(ctx, versionID, domainFiles); err != nil {
		return err
	}
	return service.repositories.Assets.ReplaceVersionAssets(ctx, versionID, assets)
}

// domainFiles maps extracted files to persisted domain files and assets.
func (service Service) domainFiles(
	ctx context.Context,
	versionID uuid.UUID,
	files []packageFile,
) ([]domain.ThemeFile, []domain.ThemeAsset, error) {
	domainFiles := make([]domain.ThemeFile, 0, len(files))
	assets := make([]domain.ThemeAsset, 0)
	for _, file := range files {
		domainFile, storageKey, err := service.domainFile(ctx, versionID, file)
		if err != nil {
			return nil, nil, err
		}
		domainFiles = append(domainFiles, domainFile)
		if file.kind == domain.FileKindAsset {
			assets = append(assets, domain.ThemeAsset{
				ID:             uuid.New(),
				VersionID:      versionID,
				FileID:         domainFile.ID,
				Path:           file.path,
				ContentType:    file.contentType,
				SizeBytes:      int64(len(file.bytes)),
				ContentSHA256:  file.sha256,
				StorageKey:     storageKey,
				IntegrityValue: "sha256-" + string(file.sha256),
			})
		}
	}
	return domainFiles, assets, nil
}

// domainFile maps and stores one package file.
func (service Service) domainFile(
	ctx context.Context,
	versionID uuid.UUID,
	file packageFile,
) (domain.ThemeFile, string, error) {
	fileID := uuid.New()
	storageKey := ""
	if file.kind == domain.FileKindAsset || !file.text {
		var err error
		storageKey, err = service.storeFile(ctx, versionID, file)
		if err != nil {
			return domain.ThemeFile{}, "", err
		}
	}
	contentText := ""
	if file.text {
		contentText = string(file.bytes)
	}
	return domain.ThemeFile{
		ID:             fileID,
		VersionID:      versionID,
		Kind:           file.kind,
		Path:           file.path,
		ContentSHA256:  file.sha256,
		ContentStorage: storageKey,
		ContentText:    contentText,
		SizeBytes:      int64(len(file.bytes)),
	}, storageKey, nil
}

// storeFile writes one immutable file to object storage.
func (service Service) storeFile(ctx context.Context, versionID uuid.UUID, file packageFile) (string, error) {
	if service.store == nil {
		return "", fmt.Errorf("theme asset storage is required")
	}
	key := storageKey(service.cfg.StoragePrefix, versionID, file.path)
	stored, err := service.store.Put(ctx, storage.Object{
		Key:         key,
		ContentType: file.contentType,
		SizeBytes:   int64(len(file.bytes)),
	}, bytes.NewReader(file.bytes))
	if err != nil {
		return "", err
	}
	return stored.Key, nil
}

// persistDiagnostics stores validation and signature results.
func (service Service) persistDiagnostics(
	ctx context.Context,
	versionID uuid.UUID,
	issues []domain.ThemeValidationIssue,
	signature domain.ThemePackageSignature,
) error {
	if err := service.repositories.Issues.ReplaceVersionIssues(ctx, versionID, issues); err != nil {
		return err
	}
	if service.repositories.Signatures == nil || signature.ID == uuid.Nil {
		return nil
	}
	return service.repositories.Signatures.ReplaceVersionSignature(ctx, versionID, signature)
}

// finalizeVersion stores final status and integrity.
func (service Service) finalizeVersion(
	ctx context.Context,
	version domain.ThemeVersion,
	files []packageFile,
	issues []domain.ThemeValidationIssue,
) (domain.ThemeVersion, error) {
	version.Status = domain.VersionStatusDraft
	if hasError(issues) {
		version.Status = domain.VersionStatusInvalid
	}
	version.IntegritySHA256 = domain.CalculateVersionIntegritySHA256(integrityFiles(files))
	return service.repositories.Versions.Update(ctx, version, version.Version)
}

// verifySignature verifies package signature files when verifier is configured.
func (service Service) verifySignature(ctx context.Context, files []packageFile, manifest []byte) signing.Result {
	if service.verifier == nil {
		return signing.Result{}
	}
	return service.verifier.Verify(ctx, manifest, findFileBytes(files, "realmkit-theme.sig.json"))
}

// sourceReference returns the stored idempotency source reference.
func sourceReference(key string) string {
	return "upload:idempotency:" + strings.TrimSpace(key)
}

// storageKey returns an immutable object key for a version file.
func storageKey(prefix string, versionID uuid.UUID, filePath domain.FilePath) string {
	return strings.Trim(prefix, "/") + "/" + versionID.String() + "/" + string(filePath)
}
