package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestMetafieldDefinitionValidateAcceptsSupportedDefinition verifies valid definitions pass.
func TestMetafieldDefinitionValidateAcceptsSupportedDefinition(t *testing.T) {
	maxLength := 80
	definition := MetafieldDefinition{
		OwnerType: OwnerUser,
		Namespace: "profile",
		Key:       "motto",
		Name:      "Motto",
		ValueType: ValueSingleLineText,
		Rules:     Rules{MaxLength: &maxLength},
		Active:    true,
	}

	if err := definition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestMetafieldDefinitionValidateRejectsUnsupportedOwner verifies owner type allowlisting.
func TestMetafieldDefinitionValidateRejectsUnsupportedOwner(t *testing.T) {
	definition := MetafieldDefinition{
		OwnerType: "raw_table_name",
		Namespace: "profile",
		Key:       "motto",
		Name:      "Motto",
		ValueType: ValueSingleLineText,
	}

	err := definition.Validate()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestFieldDefinitionValidateRejectsInvalidRuleCombination verifies rule compatibility.
func TestFieldDefinitionValidateRejectsInvalidRuleCombination(t *testing.T) {
	precision := 10
	field := FieldDefinition{
		Key:       "motto",
		Name:      "Motto",
		ValueType: ValueInteger,
		Rules:     Rules{Precision: &precision},
	}

	violations := field.Validate("field")
	if len(violations) == 0 {
		t.Fatalf("Validate() violations = 0, want at least 1")
	}
}

// TestNormalizeValueWrapsScalar verifies scalar values are stored canonically.
func TestNormalizeValueWrapsScalar(t *testing.T) {
	field := FieldDefinition{Key: "motto", Name: "Motto", ValueType: ValueSingleLineText}

	got, err := NormalizeValue(field, json.RawMessage(`"Hello"`))
	if err != nil {
		t.Fatalf("NormalizeValue() error = %v", err)
	}

	if string(got) != `{"value":"Hello"}` {
		t.Fatalf("NormalizeValue() = %s", got)
	}
}

// TestNormalizeValueRejectsSingleLineBreak verifies single-line text stays single-line.
func TestNormalizeValueRejectsSingleLineBreak(t *testing.T) {
	field := FieldDefinition{Key: "motto", Name: "Motto", ValueType: ValueSingleLineText}

	if _, err := NormalizeValue(field, json.RawMessage("\"Hello\\nworld\"")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("NormalizeValue() error = %v, want %v", err, ErrInvalid)
	}
}

// TestNormalizeValueRejectsDuplicateListItems verifies unique list rules.
func TestNormalizeValueRejectsDuplicateListItems(t *testing.T) {
	field := FieldDefinition{
		Key:       "colors",
		Name:      "Colors",
		ValueType: ValueColor,
		List:      true,
		Rules:     Rules{UniqueItems: true},
	}

	if _, err := NormalizeValue(field, json.RawMessage(`["#ffffff","#ffffff"]`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("NormalizeValue() error = %v, want %v", err, ErrInvalid)
	}
}

// TestNormalizeValueNormalizesOwnerReference verifies owner references remain structured.
func TestNormalizeValueNormalizesOwnerReference(t *testing.T) {
	id := uuid.New()
	field := FieldDefinition{
		Key:       "friend",
		Name:      "Friend",
		ValueType: ValueOwnerReference,
		Rules:     Rules{AllowedOwnerTypes: []OwnerType{OwnerUser}},
	}

	got, err := NormalizeValue(field, json.RawMessage(`{"type":"user","id":"`+id.String()+`"}`))
	if err != nil {
		t.Fatalf("NormalizeValue() error = %v", err)
	}

	if string(got) != `{"type":"user","id":"`+id.String()+`"}` {
		t.Fatalf("NormalizeValue() = %s", got)
	}
}

// TestMetaobjectDefinitionValidateRejectsDuplicateFields verifies field keys are unique.
func TestMetaobjectDefinitionValidateRejectsDuplicateFields(t *testing.T) {
	definition := MetaobjectDefinition{
		Type: "profile_card",
		Name: "Profile Card",
		Fields: []FieldDefinition{
			{Key: "motto", Name: "Motto", ValueType: ValueSingleLineText},
			{Key: "motto", Name: "Motto Again", ValueType: ValueSingleLineText},
		},
	}

	if err := definition.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestValidateMetaobjectEntryFieldsRejectsUnknownField verifies entry schema enforcement.
func TestValidateMetaobjectEntryFieldsRejectsUnknownField(t *testing.T) {
	definition := MetaobjectDefinition{
		Type: "profile_card",
		Name: "Profile Card",
		Fields: []FieldDefinition{
			{Key: "motto", Name: "Motto", ValueType: ValueSingleLineText, Required: true},
		},
	}

	_, err := ValidateMetaobjectEntryFields(definition, map[Key]json.RawMessage{
		"motto":   json.RawMessage(`"Ready"`),
		"unknown": json.RawMessage(`"nope"`),
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ValidateMetaobjectEntryFields() error = %v, want %v", err, ErrInvalid)
	}
}

// TestNormalizeValueValidationMatrix verifies representative value types.
func TestNormalizeValueValidationMatrix(t *testing.T) {
	min := 1.0
	max := 10.0
	scale := 2
	maxBytes := 32
	tests := []struct {
		name  string
		field FieldDefinition
		raw   json.RawMessage
		want  string
	}{
		{
			name:  "integer",
			field: FieldDefinition{Key: "score", Name: "Score", ValueType: ValueInteger, Rules: Rules{Min: &min, Max: &max}},
			raw:   json.RawMessage(`5`),
			want:  `{"value":5}`,
		},
		{
			name:  "decimal",
			field: FieldDefinition{Key: "weight", Name: "Weight", ValueType: ValueDecimal, Rules: Rules{Scale: &scale}},
			raw:   json.RawMessage(`"12.34"`),
			want:  `{"value":"12.34"}`,
		},
		{
			name:  "boolean",
			field: FieldDefinition{Key: "enabled", Name: "Enabled", ValueType: ValueBoolean},
			raw:   json.RawMessage(`true`),
			want:  `{"value":true}`,
		},
		{
			name:  "date",
			field: FieldDefinition{Key: "birthday", Name: "Birthday", ValueType: ValueDate},
			raw:   json.RawMessage(`"2026-06-08"`),
			want:  `{"value":"2026-06-08"}`,
		},
		{
			name:  "datetime",
			field: FieldDefinition{Key: "seen_at", Name: "Seen At", ValueType: ValueDatetime},
			raw:   json.RawMessage(`"2026-06-08T10:00:00-05:00"`),
			want:  `{"value":"2026-06-08T15:00:00Z"}`,
		},
		{
			name:  "json",
			field: FieldDefinition{Key: "payload", Name: "Payload", ValueType: ValueJSON, Rules: Rules{MaxBytes: &maxBytes}},
			raw:   json.RawMessage(`{"ok":true}`),
			want:  `{"value":{"ok":true}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeValue(test.field, test.raw)
			if err != nil {
				t.Fatalf("NormalizeValue() error = %v", err)
			}
			if string(got) != test.want {
				t.Fatalf("NormalizeValue() = %s, want %s", got, test.want)
			}
		})
	}
}

// TestNormalizeValueRejectsRequiredNull verifies required values cannot be null.
func TestNormalizeValueRejectsRequiredNull(t *testing.T) {
	field := FieldDefinition{Key: "motto", Name: "Motto", ValueType: ValueSingleLineText, Required: true}

	if _, err := NormalizeValue(field, json.RawMessage(`null`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("NormalizeValue() error = %v, want %v", err, ErrInvalid)
	}
}
