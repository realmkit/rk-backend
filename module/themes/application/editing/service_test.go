package editing

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// TestDraftFileLifecycleUsesETagsAndRevalidates verifies editor file commands.
func TestDraftFileLifecycleUsesETagsAndRevalidates(t *testing.T) {
	repositories, versionID := editingRepositories(domain.VersionStatusDraft)
	validator := &fakeValidator{}
	service := NewService(repositories, validator)
	created, err := service.CreateFile(context.Background(), WriteFileCommand{
		VersionID:   versionID,
		Path:        "templates/home.liquid",
		Kind:        domain.FileKindTemplate,
		ContentText: "home",
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("validator calls = %d, want 1", validator.calls)
	}
	if _, err := service.UpdateFile(context.Background(), WriteFileCommand{
		VersionID:    versionID,
		FileID:       created.File.ID,
		Path:         created.File.Path,
		Kind:         created.File.Kind,
		ContentText:  "changed",
		ExpectedETag: `"stale"`,
		ActorUserID:  nil,
	}); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("UpdateFile(stale) error = %v, want precondition", err)
	}
	updated, err := service.UpdateFile(context.Background(), WriteFileCommand{
		VersionID:    versionID,
		FileID:       created.File.ID,
		Path:         created.File.Path,
		Kind:         created.File.Kind,
		ContentText:  "changed",
		ExpectedETag: created.ETag,
		ActorUserID:  nil,
	})
	if err != nil {
		t.Fatalf("UpdateFile() error = %v", err)
	}
	if updated.File.ID != created.File.ID || updated.ETag == created.ETag {
		t.Fatalf("updated = %+v, want same id and new ETag", updated)
	}
	if err := service.DeleteFile(context.Background(), DeleteFileCommand{
		VersionID: versionID, FileID: updated.File.ID, ExpectedETag: updated.ETag,
	}); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
}

// TestCloneDraftCopiesFiles verifies draft creation from an existing version.
func TestCloneDraftCopiesFiles(t *testing.T) {
	repositories, versionID := editingRepositories(domain.VersionStatusPublished)
	repositories.Files.(*fakeFileRepository).files[versionID] = []domain.ThemeFile{
		{ID: uuid.New(), VersionID: versionID, Kind: domain.FileKindTemplate, Path: "templates/home.liquid"},
	}
	validator := &fakeValidator{}
	service := NewService(repositories, validator)
	cloned, err := service.CloneDraft(context.Background(), CloneDraftCommand{SourceVersionID: versionID, Label: "Work"})
	if err != nil {
		t.Fatalf("CloneDraft() error = %v", err)
	}
	files := repositories.Files.(*fakeFileRepository).files[cloned.ID]
	if cloned.Status != domain.VersionStatusDraft || len(files) != 1 || files[0].VersionID != cloned.ID {
		t.Fatalf("cloned = %+v files = %+v, want copied draft", cloned, files)
	}
	if validator.calls != 1 {
		t.Fatalf("validator calls = %d, want 1", validator.calls)
	}
}

// TestPublishedVersionRejectsMutation verifies immutable version protection.
func TestPublishedVersionRejectsMutation(t *testing.T) {
	repositories, versionID := editingRepositories(domain.VersionStatusPublished)
	service := NewService(repositories, nil)
	_, err := service.CreateFile(context.Background(), WriteFileCommand{
		VersionID: versionID, Path: "templates/home.liquid", Kind: domain.FileKindTemplate, ContentText: "home",
	})
	if !errors.Is(err, port.ErrInvalidState) {
		t.Fatalf("CreateFile(published) error = %v, want invalid state", err)
	}
}

// editingRepositories returns fake editing repositories.
func editingRepositories(status domain.VersionStatus) (Repositories, uuid.UUID) {
	versionID := uuid.New()
	return Repositories{
		Versions: &fakeVersionRepository{versions: map[uuid.UUID]domain.ThemeVersion{
			versionID: {ID: versionID, ThemeID: uuid.New(), Status: status, Label: "Source", Version: 1},
		}},
		Files:  &fakeFileRepository{files: map[uuid.UUID][]domain.ThemeFile{}},
		Assets: fakeAssetRepository{},
	}, versionID
}
