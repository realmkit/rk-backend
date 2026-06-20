package editing

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// replaceAndValidate persists the draft file snapshot and revalidates.
func (service Service) replaceAndValidate(
	ctx context.Context,
	versionID uuid.UUID,
	files []domain.ThemeFile,
	actorUserID *uuid.UUID,
) error {
	if err := service.repositories.Files.ReplaceVersionFiles(ctx, versionID, files); err != nil {
		return err
	}
	if service.validator == nil {
		return nil
	}
	_, err := service.validator.Validate(ctx, ValidationCommand{VersionID: versionID, ActorUserID: actorUserID})
	return err
}

// findFile returns one file by ID.
func (service Service) findFile(
	ctx context.Context,
	versionID uuid.UUID,
	fileID uuid.UUID,
) (domain.ThemeFile, error) {
	files, err := service.repositories.Files.ListByVersion(ctx, versionID)
	if err != nil {
		return domain.ThemeFile{}, err
	}
	for _, file := range files {
		if file.ID == fileID {
			return file, nil
		}
	}
	return domain.ThemeFile{}, port.ErrNotFound
}
