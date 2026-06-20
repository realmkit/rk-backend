package editing

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// NewService creates a draft editing service.
func NewService(repositories Repositories, validator Validator) Service {
	return Service{repositories: repositories, validator: validator}
}

// ListFiles returns files for a version.
func (service Service) ListFiles(ctx context.Context, versionID uuid.UUID) ([]FileResult, error) {
	files, err := service.repositories.Files.ListByVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}
	results := make([]FileResult, 0, len(files))
	for _, file := range files {
		results = append(results, FileResult{File: file, ETag: ETag(file)})
	}
	return results, nil
}

// GetFile returns one file and its ETag.
func (service Service) GetFile(ctx context.Context, versionID uuid.UUID, fileID uuid.UUID) (FileResult, error) {
	file, err := service.findFile(ctx, versionID, fileID)
	if err != nil {
		return FileResult{}, err
	}
	return FileResult{File: file, ETag: ETag(file)}, nil
}

// CreateFile creates a new draft file.
func (service Service) CreateFile(ctx context.Context, command WriteFileCommand) (FileResult, error) {
	if err := service.ensureDraft(ctx, command.VersionID); err != nil {
		return FileResult{}, err
	}
	files, err := service.repositories.Files.ListByVersion(ctx, command.VersionID)
	if err != nil {
		return FileResult{}, err
	}
	file, err := newFile(command)
	if err != nil {
		return FileResult{}, err
	}
	files = append(files, file)
	if err := service.replaceAndValidate(ctx, command.VersionID, files, command.ActorUserID); err != nil {
		return FileResult{}, err
	}
	return FileResult{File: file, ETag: ETag(file)}, nil
}

// UpdateFile updates an existing draft file.
func (service Service) UpdateFile(ctx context.Context, command WriteFileCommand) (FileResult, error) {
	if err := service.ensureDraft(ctx, command.VersionID); err != nil {
		return FileResult{}, err
	}
	files, err := service.repositories.Files.ListByVersion(ctx, command.VersionID)
	if err != nil {
		return FileResult{}, err
	}
	next, err := updateFile(files, command)
	if err != nil {
		return FileResult{}, err
	}
	if err := service.replaceAndValidate(ctx, command.VersionID, files, command.ActorUserID); err != nil {
		return FileResult{}, err
	}
	return FileResult{File: next, ETag: ETag(next)}, nil
}

// DeleteFile deletes an existing draft file.
func (service Service) DeleteFile(ctx context.Context, command DeleteFileCommand) error {
	if err := service.ensureDraft(ctx, command.VersionID); err != nil {
		return err
	}
	files, err := service.repositories.Files.ListByVersion(ctx, command.VersionID)
	if err != nil {
		return err
	}
	files, err = deleteFile(files, command)
	if err != nil {
		return err
	}
	return service.replaceAndValidate(ctx, command.VersionID, files, command.ActorUserID)
}

// CloneDraft creates a new editable draft from another version.
func (service Service) CloneDraft(ctx context.Context, command CloneDraftCommand) (domain.ThemeVersion, error) {
	source, err := service.repositories.Versions.FindByID(ctx, command.SourceVersionID)
	if err != nil {
		return domain.ThemeVersion{}, err
	}
	files, err := service.repositories.Files.ListByVersion(ctx, source.ID)
	if err != nil {
		return domain.ThemeVersion{}, err
	}
	version, err := service.repositories.Versions.Create(ctx, cloneVersion(source, command))
	if err != nil {
		return domain.ThemeVersion{}, err
	}
	if err := service.repositories.Files.ReplaceVersionFiles(ctx, version.ID, cloneFiles(version.ID, files)); err != nil {
		return domain.ThemeVersion{}, err
	}
	if service.validator != nil {
		if _, err := service.validator.Validate(ctx, ValidationCommand{VersionID: version.ID, ActorUserID: command.ActorUserID}); err != nil {
			return domain.ThemeVersion{}, err
		}
	}
	return service.repositories.Versions.FindByID(ctx, version.ID)
}

// ensureDraft verifies a version can be edited.
func (service Service) ensureDraft(ctx context.Context, versionID uuid.UUID) error {
	version, err := service.repositories.Versions.FindByID(ctx, versionID)
	if err != nil {
		return err
	}
	if version.Status != domain.VersionStatusDraft && version.Status != domain.VersionStatusInvalid {
		return port.ErrInvalidState
	}
	return version.EnsureEditable()
}
