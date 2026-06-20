package editing

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// cloneVersion returns a new draft version from a source version.
func cloneVersion(source domain.ThemeVersion, command CloneDraftCommand) domain.ThemeVersion {
	label := command.Label
	if label == "" {
		label = source.Label + " draft"
	}
	return domain.ThemeVersion{
		ID:                 uuid.New(),
		ThemeID:            source.ThemeID,
		Label:              label,
		Status:             domain.VersionStatusDraft,
		SourceKind:         domain.SourceEditor,
		SourceReference:    "draft-from:" + source.ID.String(),
		ManifestJSON:       source.ManifestJSON,
		SettingsSchemaJSON: source.SettingsSchemaJSON,
		SettingsDataJSON:   source.SettingsDataJSON,
		IntegritySHA256:    source.IntegritySHA256,
		CreatedBy:          command.ActorUserID,
		UpdatedBy:          command.ActorUserID,
		Version:            1,
	}
}

// cloneFiles returns files copied into a new version.
func cloneFiles(versionID uuid.UUID, files []domain.ThemeFile) []domain.ThemeFile {
	copied := make([]domain.ThemeFile, 0, len(files))
	for _, file := range files {
		file.ID = uuid.New()
		file.VersionID = versionID
		copied = append(copied, file)
	}
	return copied
}
