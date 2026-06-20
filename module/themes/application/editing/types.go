package editing

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by draft editing.
type Repositories struct {
	Versions port.VersionRepository
	Files    port.FileRepository
	Assets   port.AssetRepository
}

// Validator revalidates a version after a draft file mutation.
type Validator interface {
	Validate(context.Context, ValidationCommand) (ValidationResult, error)
}

// ValidationCommand is the subset required from the validation use case.
type ValidationCommand struct {
	VersionID   uuid.UUID
	ActorUserID *uuid.UUID
}

// ValidationResult is the subset returned from validation.
type ValidationResult struct {
	Version domain.ThemeVersion
	Issues  []domain.ThemeValidationIssue
}

// Service edits draft theme files.
type Service struct {
	repositories Repositories
	validator    Validator
}

// FileResult contains one file and its ETag.
type FileResult struct {
	File domain.ThemeFile
	ETag string
}

// WriteFileCommand creates or updates a draft file.
type WriteFileCommand struct {
	VersionID    uuid.UUID
	FileID       uuid.UUID
	Path         domain.FilePath
	Kind         domain.FileKind
	ContentText  string
	ExpectedETag string
	ActorUserID  *uuid.UUID
}

// DeleteFileCommand deletes a draft file.
type DeleteFileCommand struct {
	VersionID    uuid.UUID
	FileID       uuid.UUID
	ExpectedETag string
	ActorUserID  *uuid.UUID
}

// CloneDraftCommand creates a draft version from another version.
type CloneDraftCommand struct {
	SourceVersionID uuid.UUID
	Label           string
	ActorUserID     *uuid.UUID
}
