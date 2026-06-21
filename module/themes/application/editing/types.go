package editing

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// Repositories contains persistence ports required by draft editing.
type Repositories struct {
	Versions port.VersionRepository // Versions stores the versions value.
	Files    port.FileRepository    // Files stores the files value.
	Assets   port.AssetRepository   // Assets stores the assets value.
}

// Validator revalidates a version after a draft file mutation.
type Validator interface {
	Validate(context.Context, ValidationCommand) (ValidationResult, error)
}

// ValidationCommand is the subset required from the validation use case.
type ValidationCommand struct {
	VersionID   uuid.UUID  // VersionID stores the version i d value.
	ActorUserID *uuid.UUID // ActorUserID stores the actor user i d value.
}

// ValidationResult is the subset returned from validation.
type ValidationResult struct {
	Version domain.ThemeVersion           // Version stores the version value.
	Issues  []domain.ThemeValidationIssue // Issues stores the issues value.
}

// Service edits draft theme files.
type Service struct {
	repositories Repositories // repositories stores the repositories value.
	validator    Validator    // validator stores the validator value.
}

// FileResult contains one file and its ETag.
type FileResult struct {
	File domain.ThemeFile // File stores the file value.
	ETag string           // ETag stores the e tag value.
}

// WriteFileCommand creates or updates a draft file.
type WriteFileCommand struct {
	VersionID    uuid.UUID       // VersionID stores the version i d value.
	FileID       uuid.UUID       // FileID stores the file i d value.
	Path         domain.FilePath // Path stores the path value.
	Kind         domain.FileKind // Kind stores the kind value.
	ContentText  string          // ContentText stores the content text value.
	ExpectedETag string          // ExpectedETag stores the expected e tag value.
	ActorUserID  *uuid.UUID      // ActorUserID stores the actor user i d value.
}

// DeleteFileCommand deletes a draft file.
type DeleteFileCommand struct {
	VersionID    uuid.UUID  // VersionID stores the version i d value.
	FileID       uuid.UUID  // FileID stores the file i d value.
	ExpectedETag string     // ExpectedETag stores the expected e tag value.
	ActorUserID  *uuid.UUID // ActorUserID stores the actor user i d value.
}

// CloneDraftCommand creates a draft version from another version.
type CloneDraftCommand struct {
	SourceVersionID uuid.UUID  // SourceVersionID stores the source version i d value.
	Label           string     // Label stores the label value.
	ActorUserID     *uuid.UUID // ActorUserID stores the actor user i d value.
}
