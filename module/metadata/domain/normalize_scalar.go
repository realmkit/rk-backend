package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// normalizeString validates and normalizes string values.
func normalizeString(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be a string"}})
	}
	var violations []Violation
	if field.ValueType == ValueSingleLineText && strings.ContainsAny(value, "\r\n") {
		violations = AppendViolation(violations, "value", "must not contain line breaks")
	}
	if field.Rules.MinLength != nil && len(value) < *field.Rules.MinLength {
		violations = AppendViolation(violations, "value", "is shorter than min_length")
	}
	if field.Rules.MaxLength != nil && len(value) > *field.Rules.MaxLength {
		violations = AppendViolation(violations, "value", "is longer than max_length")
	}
	if field.Rules.Pattern != "" && !regexp.MustCompile(field.Rules.Pattern).MatchString(value) {
		violations = AppendViolation(violations, "value", "does not match pattern")
	}
	violations = append(violations, validateSpecialString(field, value)...)
	return marshalItem(value, violations)
}

// validateSpecialString validates string subtypes.
func validateSpecialString(field FieldDefinition, value string) []Violation {
	var violations []Violation
	if field.ValueType == ValueURL {
		parsed, err := url.Parse(value)
		if err != nil || !isHTTPURL(parsed) {
			violations = AppendViolation(violations, "value", "must be an absolute http or https URL")
		}
	}
	if field.ValueType == ValueColor && !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(value) {
		violations = AppendViolation(violations, "value", "must be a hex color")
	}
	if field.ValueType == ValueEnum && !slices.Contains(field.Rules.AllowedValues, value) {
		violations = AppendViolation(violations, "value", "must be an allowed value")
	}
	return violations
}

// isHTTPURL reports whether parsed is an absolute HTTP(S) URL.
func isHTTPURL(parsed *url.URL) bool {
	return parsed.Scheme != "" && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

// normalizeInteger validates and normalizes integer values.
func normalizeInteger(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var number json.Number
	if err := decodeNumber(raw, &number); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be an integer"}})
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be an integer"}})
	}
	var violations []Violation
	if field.Rules.Min != nil && float64(value) < *field.Rules.Min {
		violations = AppendViolation(violations, "value", "is less than min")
	}
	if field.Rules.Max != nil && float64(value) > *field.Rules.Max {
		violations = AppendViolation(violations, "value", "is greater than max")
	}
	return marshalItem(value, violations)
}

// normalizeDecimal validates and normalizes decimal values.
func normalizeDecimal(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	value := strings.Trim(string(raw), `"`)
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsInf(floatValue, 0) || math.IsNaN(floatValue) {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be a decimal"}})
	}
	var violations []Violation
	if field.Rules.Min != nil && floatValue < *field.Rules.Min {
		violations = AppendViolation(violations, "value", "is less than min")
	}
	if field.Rules.Max != nil && floatValue > *field.Rules.Max {
		violations = AppendViolation(violations, "value", "is greater than max")
	}
	if field.Rules.Scale != nil && decimalScale(value) > *field.Rules.Scale {
		violations = AppendViolation(violations, "value", "has too many decimal places")
	}
	return marshalItem(value, violations)
}

// normalizeBoolean validates and normalizes boolean values.
func normalizeBoolean(raw json.RawMessage) (json.RawMessage, error) {
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be a boolean"}})
	}
	return marshalCanonical(value)
}

// normalizeDate validates and normalizes date values.
func normalizeDate(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be a date string"}})
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must use YYYY-MM-DD"}})
	}
	return marshalItem(parsed.Format(time.DateOnly), validateTemporalValue(field, parsed))
}

// normalizeDatetime validates and normalizes datetime values.
func normalizeDatetime(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be a datetime string"}})
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must use RFC3339"}})
	}
	return marshalItem(parsed.UTC().Format(time.RFC3339), validateTemporalValue(field, parsed))
}

// normalizeJSON validates and normalizes arbitrary JSON values.
func normalizeJSON(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be valid JSON"}})
	}
	if field.Rules.MaxBytes != nil && len(raw) > *field.Rules.MaxBytes {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "exceeds max_bytes"}})
	}
	return marshalCanonical(value)
}

// normalizeOwnerReference validates and normalizes owner references.
func normalizeOwnerReference(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var value OwnerReference
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be an owner reference"}})
	}
	var violations []Violation
	violations = append(violations, ValidateOwnerType("value.type", value.Type)...)
	if value.ID == uuid.Nil {
		violations = AppendViolation(violations, "value.id", "is required")
	}
	if len(field.Rules.AllowedOwnerTypes) > 0 && !slices.Contains(field.Rules.AllowedOwnerTypes, value.Type) {
		violations = AppendViolation(violations, "value.type", "is not allowed")
	}
	return marshalItem(value, violations)
}

// normalizeMetaobjectReference validates and normalizes metaobject references.
func normalizeMetaobjectReference(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var value MetaobjectReference
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be a metaobject reference"}})
	}
	var violations []Violation
	if value.DefinitionID == uuid.Nil {
		violations = AppendViolation(violations, "value.definition_id", "is required")
	}
	if value.EntryID == uuid.Nil {
		violations = AppendViolation(violations, "value.entry_id", "is required")
	}
	if len(field.Rules.MetaobjectDefinitionIDs) > 0 &&
		!containsUUID(field.Rules.MetaobjectDefinitionIDs, value.DefinitionID) {
		violations = AppendViolation(violations, "value.definition_id", "is not allowed")
	}
	return marshalItem(value, violations)
}

// containsUUID reports whether values contains value.
func containsUUID(values []uuid.UUID, value uuid.UUID) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

// validateTemporalValue validates temporal constraints.
func validateTemporalValue(field FieldDefinition, value time.Time) []Violation {
	var violations []Violation
	if field.Rules.FutureOnly && !value.After(time.Now()) {
		violations = AppendViolation(violations, "value", "must be in the future")
	}
	if field.Rules.PastOnly && !value.Before(time.Now()) {
		violations = AppendViolation(violations, "value", "must be in the past")
	}
	return violations
}

// decodeNumber decodes raw as a JSON number.
func decodeNumber(raw json.RawMessage, number *json.Number) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(number); err != nil {
		return err
	}
	if strings.Contains(number.String(), ".") {
		return fmt.Errorf("decimal number: %s", number.String())
	}
	return nil
}

// decimalScale returns the number of digits after the decimal point.
func decimalScale(value string) int {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return 0
	}
	return len(parts[1])
}

// marshalItem marshals value when violations is empty.
func marshalItem(value any, violations []Violation) (json.RawMessage, error) {
	if err := NewValidationError(violations); err != nil {
		return nil, err
	}
	return marshalCanonical(value)
}

// marshalCanonical marshals a canonical JSON value.
func marshalCanonical(value any) (json.RawMessage, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
