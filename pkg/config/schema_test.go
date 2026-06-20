package config

import (
	"reflect"
	"testing"
)

// TestSchemaRequiresFieldsWithoutDefaults verifies a field without a default tag is mandatory.
func TestSchemaRequiresFieldsWithoutDefaults(t *testing.T) {
	clearRealmKitEnv(t)

	type requiredConfig struct {
		Token string `mapstructure:"token"`
	}

	fields, err := schemaFor(requiredConfig{})
	if err != nil {
		t.Fatalf("schemaFor() error = %v", err)
	}
	source := newViper(defaultPrefix)
	for _, field := range fields {
		if err := source.BindEnv(field.key, field.env(defaultPrefix)); err != nil {
			t.Fatalf("BindEnv() error = %v", err)
		}
	}
	if err := validateRequired(source, fields); err == nil {
		t.Fatalf("validateRequired() error = nil, want error")
	}
}

// TestValidateRequiredRejectsEmptyString verifies mandatory string settings cannot be blank.
func TestValidateRequiredRejectsEmptyString(t *testing.T) {
	fields := []fieldSpec{{key: "token"}}
	source := newViper(defaultPrefix)
	source.Set("token", " ")
	if err := validateRequired(source, fields); err == nil {
		t.Fatalf("validateRequired() error = nil, want error")
	}
}

// TestValidateRequiredAcceptsPresentValues verifies mandatory settings pass when configured.
func TestValidateRequiredAcceptsPresentValues(t *testing.T) {
	fields := []fieldSpec{{key: "enabled"}}
	source := newViper(defaultPrefix)
	source.Set("enabled", false)
	if err := validateRequired(source, fields); err != nil {
		t.Fatalf("validateRequired() error = %v", err)
	}
}

// TestSchemaCollectsSquashedFields verifies squashed structs expose root-level REALMKIT variables.
func TestSchemaCollectsSquashedFields(t *testing.T) {
	fields, err := schemaFor(Config{})
	if err != nil {
		t.Fatalf("schemaFor() error = %v", err)
	}
	got := make([]string, 0, len(fields))
	for _, field := range fields {
		got = append(got, field.key)
	}
	want := rootConfigFieldKeys()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
}

// TestRuntimeIsDevelopment verifies development environment matching is normalized.
func TestRuntimeIsDevelopment(t *testing.T) {
	if !(Runtime{Environment: " Development "}).IsDevelopment() {
		t.Fatalf("IsDevelopment() = false, want true")
	}
	if (Runtime{Environment: "production"}).IsDevelopment() {
		t.Fatalf("IsDevelopment() = true, want false")
	}
}

// TestSchemaRejectsNonStructs verifies only struct values can define configuration schemas.
func TestSchemaRejectsNonStructs(t *testing.T) {
	if _, err := schemaFor("invalid"); err == nil {
		t.Fatalf("schemaFor() error = nil, want error")
	}
}

// TestSchemaCollectsNestedAndSkippedFields verifies nested structs, skipped tags, and fallback names.
func TestSchemaCollectsNestedAndSkippedFields(t *testing.T) {
	type databaseConfig struct {
		URL     string `mapstructure:"url" default:"postgres://localhost/realmkit"`
		NoTag   string `                   default:"fallback"`
		Ignored string `mapstructure:"-"`
	}
	type appConfig struct {
		Database databaseConfig `mapstructure:"database"`
	}
	fields, err := schemaFor(appConfig{})
	if err != nil {
		t.Fatalf("schemaFor() error = %v", err)
	}
	got := make([]string, 0, len(fields))
	for _, field := range fields {
		got = append(got, field.key)
	}
	want := []string{"database.url", "database.no_tag"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
	if fields[0].env(defaultPrefix) != "REALMKIT_DATABASE_URL" {
		t.Fatalf("env = %q, want %q", fields[0].env(defaultPrefix), "REALMKIT_DATABASE_URL")
	}
}
