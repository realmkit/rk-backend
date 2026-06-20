package publication

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// ensurePublishable verifies version validation and signature gates.
func (service Service) ensurePublishable(ctx context.Context, version domain.ThemeVersion) error {
	if version.Status != domain.VersionStatusValid && version.Status != domain.VersionStatusPublished {
		return port.ErrInvalidState
	}
	if err := service.ensureNoBlockingIssues(ctx, version.ID); err != nil {
		return err
	}
	signature, err := service.repositories.Signatures.FindByVersion(ctx, version.ID)
	if err != nil {
		return port.ErrInvalidState
	}
	if signature.VerificationStatus != domain.SignatureVerified &&
		signature.VerificationStatus != domain.SignatureRetired {
		return port.ErrInvalidState
	}
	return nil
}

// ensureNoBlockingIssues blocks versions with error diagnostics.
func (service Service) ensureNoBlockingIssues(ctx context.Context, versionID uuid.UUID) error {
	issues, err := service.repositories.Issues.ListByVersion(ctx, versionID)
	if err != nil {
		return err
	}
	for _, issue := range issues {
		if issue.Severity == domain.SeverityError {
			return port.ErrInvalidState
		}
	}
	return nil
}

// markPublished marks a valid version as publicly published.
func (service Service) markPublished(
	ctx context.Context,
	version domain.ThemeVersion,
	actorUserID *uuid.UUID,
) error {
	if version.Status == domain.VersionStatusPublished {
		return nil
	}
	now := service.clock().UTC()
	version.Status = domain.VersionStatusPublished
	version.PublishedAt = &now
	version.PublishedBy = actorUserID
	version.UpdatedBy = actorUserID
	_, err := service.repositories.Versions.Update(ctx, version, version.Version)
	return err
}
