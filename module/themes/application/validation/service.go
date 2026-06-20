package validation

import (
	"context"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// NewService creates a static validation service.
func NewService(repositories Repositories) Service {
	return Service{repositories: repositories}
}

// Validate validates one version and persists diagnostics.
func (service Service) Validate(ctx context.Context, command Command) (Result, error) {
	version, err := service.repositories.Versions.FindByID(ctx, command.VersionID)
	if err != nil {
		return Result{}, err
	}
	files, err := service.repositories.Files.ListByVersion(ctx, command.VersionID)
	if err != nil {
		return Result{}, err
	}
	issues, err := validateVersionFiles(ctx, version, files)
	if err != nil {
		return Result{}, err
	}
	manifest, err := compiledManifest(ctx, version, files, issues)
	if err != nil {
		return Result{}, err
	}
	version.ManifestJSON = manifest
	version.SettingsSchemaJSON = settingsSchemaJSON(files)
	version.SettingsDataJSON = settingsDataJSON(files)
	version.IntegritySHA256 = domain.CalculateVersionIntegritySHA256(integrityFiles(files))
	version.Status = validationStatus(issues)
	version.UpdatedBy = command.ActorUserID
	if err := service.repositories.Issues.ReplaceVersionIssues(ctx, version.ID, issues); err != nil {
		return Result{}, err
	}
	updated, err := service.repositories.Versions.Update(ctx, version, version.Version)
	if err != nil {
		return Result{}, err
	}
	return Result{Version: updated, Issues: issues, ManifestJSON: manifest}, nil
}

// validateVersionFiles returns all static validation issues.
func validateVersionFiles(ctx context.Context, version domain.ThemeVersion, files []domain.ThemeFile) ([]domain.ThemeValidationIssue, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	index := indexFiles(files)
	issues := make([]domain.ThemeValidationIssue, 0)
	issues = append(issues, validateRequiredStructure(index)...)
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	issues = append(issues, validateJSONFiles(ctx, version, files)...)
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	issues = append(issues, validateLiquidFiles(ctx, files, index)...)
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	issues = append(issues, validateCSSFiles(ctx, files)...)
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	issues = append(issues, validateJavaScriptFiles(ctx, files)...)
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	issues = append(issues, validateRouteCoverage(ctx, index)...)
	return issues, nil
}

// validationStatus returns the final version status for issues.
func validationStatus(issues []domain.ThemeValidationIssue) domain.VersionStatus {
	for _, issue := range issues {
		if issue.Severity == domain.SeverityError {
			return domain.VersionStatusInvalid
		}
	}
	return domain.VersionStatusValid
}

// integrityFiles maps stored files to integrity inputs.
func integrityFiles(files []domain.ThemeFile) []domain.IntegrityFile {
	inputs := make([]domain.IntegrityFile, 0, len(files))
	for _, file := range files {
		inputs = append(inputs, domain.IntegrityFile{Path: file.Path, ContentSHA256: file.ContentSHA256})
	}
	return inputs
}

// checkContext returns ctx cancellation when present.
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
