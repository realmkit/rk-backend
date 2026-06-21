package publication

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// ensurePublishable verifies version validation and signature gates.
func (service Service) ensurePublishable(ctx context.Context, version domain.ThemeVersion) error {
	issues, err := service.repositories.Issues.ListByVersion(ctx, version.ID)
	if err != nil {
		return err
	}
	signature, err := service.repositories.Signatures.FindByVersion(ctx, version.ID)
	if err != nil {
		return port.ErrInvalidState
	}
	if err := version.EnsurePublishable(signature, issues); err != nil {
		return port.ErrInvalidState
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
	published, err := version.MarkPublished(service.clock(), actorUserID)
	if err != nil {
		return port.ErrInvalidState
	}
	_, err = service.repositories.Versions.Update(ctx, published, version.Version)
	return err
}
