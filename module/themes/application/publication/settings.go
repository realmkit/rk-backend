package publication

import (
	"encoding/json"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// activationSettings returns validated activation settings.
func activationSettings(version domain.ThemeVersion, value []byte) ([]byte, error) {
	if len(value) == 0 {
		value = version.SettingsDataJSON
	}
	if len(value) == 0 {
		value = []byte(`{}`)
	}
	schema, err := decodeObject(version.SettingsSchemaJSON)
	if err != nil {
		return nil, err
	}
	settings, err := decodeObject(value)
	if err != nil {
		return nil, err
	}
	if err := validateRequiredSettings(schema, settings); err != nil {
		return nil, err
	}
	return json.Marshal(settings)
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

// validateRequiredSettings verifies required settings keys.
func validateRequiredSettings(schema map[string]any, settings map[string]any) error {
	required, ok := schema["required"].([]any)
	if !ok {
		return nil
	}
	for _, raw := range required {
		key, ok := raw.(string)
		if ok {
			if _, exists := settings[key]; !exists {
				return port.ErrInvalidState
			}
		}
	}
	return nil
}
