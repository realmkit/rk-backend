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

// TestAllowedOwnerTypesIncludeForumOwners verifies forum metadata ownership.
func TestAllowedOwnerTypesIncludeForumOwners(t *testing.T) {
	for _, ownerType := range []OwnerType{OwnerForumCategory, OwnerForum, OwnerForumThread} {
		if violations := ValidateOwnerType("owner_type", ownerType); len(violations) != 0 {
			t.Fatalf("ValidateOwnerType(%q) violations = %+v, want none", ownerType, violations)
		}
	}
}

// TestAllowedOwnerTypesIncludeTicketOwners verifies ticket metadata ownership.
func TestAllowedOwnerTypesIncludeTicketOwners(t *testing.T) {
	for _, ownerType := range []OwnerType{OwnerTicketDefinition, OwnerTicket} {
		if violations := ValidateOwnerType("owner_type", ownerType); len(violations) != 0 {
			t.Fatalf("ValidateOwnerType(%q) violations = %+v, want none", ownerType, violations)
		}
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

// TestRulesValidateReportsInvalidCombinations verifies rule consistency checks.
func TestRulesValidateReportsInvalidCombinations(t *testing.T) {
	negative := -1
	one := 1
	two := 2
	min := 10.0
	max := 1.0
	definitionID := uuid.New()
	tests := []struct {
		name      string
		rules     Rules
		valueType ValueType
		list      bool
	}{
		{name: "list bounds", rules: Rules{MinItems: &two, MaxItems: &one}, valueType: ValueSingleLineText, list: true},
		{name: "non list rules", rules: Rules{MinItems: &one, UniqueItems: true}, valueType: ValueSingleLineText},
		{name: "string bounds", rules: Rules{MinLength: &two, MaxLength: &one}, valueType: ValueSingleLineText},
		{name: "negative length", rules: Rules{MinLength: &negative, MaxLength: &negative}, valueType: ValueSingleLineText},
		{name: "enum values", rules: Rules{}, valueType: ValueEnum},
		{name: "number bounds", rules: Rules{Min: &min, Max: &max}, valueType: ValueInteger},
		{name: "precision bounds", rules: Rules{Precision: &negative, Scale: &two}, valueType: ValueDecimal},
		{name: "non decimal precision", rules: Rules{Precision: &one}, valueType: ValueInteger},
		{name: "temporal direction", rules: Rules{FutureOnly: true, PastOnly: true}, valueType: ValueDate},
		{name: "json max bytes", rules: Rules{MaxBytes: &negative}, valueType: ValueJSON},
		{name: "non json max bytes", rules: Rules{MaxBytes: &one}, valueType: ValueSingleLineText},
		{name: "owner rules type", rules: Rules{AllowedOwnerTypes: []OwnerType{OwnerUser}}, valueType: ValueSingleLineText},
		{name: "owner unsupported", rules: Rules{AllowedOwnerTypes: []OwnerType{"server"}}, valueType: ValueOwnerReference},
		{name: "metaobject rules type", rules: Rules{MetaobjectDefinitionIDs: []uuid.UUID{definitionID}}, valueType: ValueSingleLineText},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			violations := test.rules.Validate("rules", test.valueType, test.list)
			if len(violations) == 0 {
				t.Fatalf("Validate() violations = 0, want at least 1")
			}
		})
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

// TestNormalizeValueNormalizesMetaobjectReference verifies metaobject references remain structured.
func TestNormalizeValueNormalizesMetaobjectReference(t *testing.T) {
	definitionID := uuid.New()
	entryID := uuid.New()
	field := FieldDefinition{
		Key:       "card",
		Name:      "Card",
		ValueType: ValueMetaobjectReference,
		Rules:     Rules{MetaobjectDefinitionIDs: []uuid.UUID{definitionID}},
	}

	got, err := NormalizeValue(field, json.RawMessage(`{"definition_id":"`+definitionID.String()+`","entry_id":"`+entryID.String()+`"}`))
	if err != nil {
		t.Fatalf("NormalizeValue() error = %v", err)
	}

	want := `{"definition_id":"` + definitionID.String() + `","entry_id":"` + entryID.String() + `"}`
	if string(got) != want {
		t.Fatalf("NormalizeValue() = %s, want %s", got, want)
	}
}

// TestNormalizeValueRejectsDisallowedMetaobjectReference verifies reference allowlists.
func TestNormalizeValueRejectsDisallowedMetaobjectReference(t *testing.T) {
	field := FieldDefinition{
		Key:       "card",
		Name:      "Card",
		ValueType: ValueMetaobjectReference,
		Rules:     Rules{MetaobjectDefinitionIDs: []uuid.UUID{uuid.New()}},
	}

	_, err := NormalizeValue(field, json.RawMessage(`{"definition_id":"`+uuid.NewString()+`","entry_id":"`+uuid.NewString()+`"}`))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("NormalizeValue() error = %v, want %v", err, ErrInvalid)
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

// TestMetaobjectEntryValidateAcceptsValidEntry verifies handle validation accepts stable handles.
func TestMetaobjectEntryValidateAcceptsValidEntry(t *testing.T) {
	entry := MetaobjectEntry{Handle: "first-card", DisplayName: "First Card"}

	if err := entry.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestMetaobjectEntryValidateRejectsInvalidHandle verifies handle validation rejects unstable handles.
func TestMetaobjectEntryValidateRejectsInvalidHandle(t *testing.T) {
	entry := MetaobjectEntry{Handle: "Bad Handle", DisplayName: "First Card"}

	if err := entry.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
	}
}

// TestValidateHandleRejectsEmptyHandle verifies handle requirement errors.
func TestValidateHandleRejectsEmptyHandle(t *testing.T) {
	violations := ValidateHandle("handle", "")
	if len(violations) != 1 {
		t.Fatalf("ValidateHandle() violations = %d, want 1", len(violations))
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

// TestNormalizeValueRejectsInvalidValueMatrix verifies type-specific failures.
func TestNormalizeValueRejectsInvalidValueMatrix(t *testing.T) {
	min := 10.0
	max := 1.0
	scale := 1
	maxBytes := 2
	tests := []struct {
		name  string
		field FieldDefinition
		raw   json.RawMessage
	}{
		{name: "url", field: FieldDefinition{Key: "website", Name: "Website", ValueType: ValueURL}, raw: json.RawMessage(`"ftp://example.com"`)},
		{name: "color", field: FieldDefinition{Key: "color", Name: "Color", ValueType: ValueColor}, raw: json.RawMessage(`"#fff"`)},
		{name: "enum", field: FieldDefinition{Key: "rank", Name: "Rank", ValueType: ValueEnum, Rules: Rules{AllowedValues: []string{"admin"}}}, raw: json.RawMessage(`"member"`)},
		{name: "integer min", field: FieldDefinition{Key: "score", Name: "Score", ValueType: ValueInteger, Rules: Rules{Min: &min}}, raw: json.RawMessage(`1`)},
		{name: "integer max", field: FieldDefinition{Key: "score", Name: "Score", ValueType: ValueInteger, Rules: Rules{Max: &max}}, raw: json.RawMessage(`2`)},
		{name: "decimal scale", field: FieldDefinition{Key: "weight", Name: "Weight", ValueType: ValueDecimal, Rules: Rules{Scale: &scale}}, raw: json.RawMessage(`"1.23"`)},
		{name: "json bytes", field: FieldDefinition{Key: "payload", Name: "Payload", ValueType: ValueJSON, Rules: Rules{MaxBytes: &maxBytes}}, raw: json.RawMessage(`{"ok":true}`)},
		{name: "owner id", field: FieldDefinition{Key: "friend", Name: "Friend", ValueType: ValueOwnerReference}, raw: json.RawMessage(`{"type":"user","id":"00000000-0000-0000-0000-000000000000"}`)},
		{name: "metaobject ids", field: FieldDefinition{Key: "card", Name: "Card", ValueType: ValueMetaobjectReference}, raw: json.RawMessage(`{"definition_id":"00000000-0000-0000-0000-000000000000","entry_id":"00000000-0000-0000-0000-000000000000"}`)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NormalizeValue(test.field, test.raw); !errors.Is(err, ErrInvalid) {
				t.Fatalf("NormalizeValue() error = %v, want %v", err, ErrInvalid)
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

// TestValidationErrorFormatsViolations verifies validation summaries include field paths.
func TestValidationErrorFormatsViolations(t *testing.T) {
	err := ValidationError{Violations: []Violation{{Field: "field", Message: "is required"}}}

	if err.Error() != "invalid metadata: field: is required" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

// TestValidationErrorFormatsEmptyViolations verifies empty validation errors use the base message.
func TestValidationErrorFormatsEmptyViolations(t *testing.T) {
	err := ValidationError{}

	if err.Error() != ErrInvalid.Error() {
		t.Fatalf("Error() = %q, want %q", err.Error(), ErrInvalid.Error())
	}
}
