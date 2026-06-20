package validation

import (
	"context"
	"encoding/json"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// validateJSONFiles validates settings and locale JSON files.
func validateJSONFiles(ctx context.Context, version domain.ThemeVersion, files []domain.ThemeFile) []domain.ThemeValidationIssue {
	issues := make([]domain.ThemeValidationIssue, 0)
	if _, err := decodeObject(settingsSchemaJSON(files)); err != nil {
		issues = append(issues, issue(domain.IssueInvalidSettingsSchema, "config/settings_schema.json", "Settings schema JSON is invalid."))
	}
	settings, err := decodeObject(settingsDataJSON(files))
	if err != nil {
		issues = append(issues, issue(domain.IssueInvalidSettingsData, "config/settings_data.json", "Settings data JSON is invalid."))
	}
	if schema, err := decodeObject(settingsSchemaJSON(files)); err == nil {
		issues = append(issues, validateRequiredSettings(schema, settings)...)
	}
	for _, file := range files {
		if err := checkContext(ctx); err != nil {
			return issues
		}
		if file.Kind == domain.FileKindLocale {
			if _, err := decodeObject([]byte(file.ContentText)); err != nil {
				issues = append(issues, issue(domain.IssueInvalidLocale, file.Path, "Locale JSON is invalid."))
			}
		}
	}
	if _, err := decodeObject(version.ManifestJSON); err != nil {
		issues = append(issues, issue(domain.IssueInvalidManifest, "realmkit-theme.json", "Compiled manifest JSON is invalid."))
	}
	return issues
}

// settingsSchemaJSON returns the package settings schema.
func settingsSchemaJSON(files []domain.ThemeFile) []byte {
	return fileContentOrDefault(files, "config/settings_schema.json")
}

// settingsDataJSON returns the package default settings data.
func settingsDataJSON(files []domain.ThemeFile) []byte {
	return fileContentOrDefault(files, "config/settings_data.json")
}

// fileContentOrDefault returns file text or an empty JSON object.
func fileContentOrDefault(files []domain.ThemeFile, filePath domain.FilePath) []byte {
	for _, file := range files {
		if domain.NormalizeFilePath(file.Path) == domain.NormalizeFilePath(filePath) {
			return []byte(file.ContentText)
		}
	}
	return []byte(`{}`)
}

// decodeObject parses a JSON object.
func decodeObject(value []byte) (map[string]any, error) {
	if len(value) == 0 {
		value = []byte(`{}`)
	}
	var object map[string]any
	if err := json.Unmarshal(value, &object); err != nil {
		return nil, err
	}
	if object == nil {
		object = map[string]any{}
	}
	return object, nil
}

// validateRequiredSettings verifies simple required settings schema entries.
func validateRequiredSettings(schema map[string]any, settings map[string]any) []domain.ThemeValidationIssue {
	required, ok := schema["required"].([]any)
	if !ok {
		return nil
	}
	issues := make([]domain.ThemeValidationIssue, 0)
	for _, raw := range required {
		key, ok := raw.(string)
		if !ok {
			continue
		}
		if _, exists := settings[key]; !exists {
			issues = append(issues, issue(domain.IssueInvalidSettingsData, "config/settings_data.json", "Required setting is missing."))
		}
	}
	return issues
}
