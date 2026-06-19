package openapi

import (
	"encoding/json"
	"slices"
	"testing"
)

// TestThemeContractDeclaresMilestoneZeroOperations verifies theme API surface is contracted.
func TestThemeContractDeclaresMilestoneZeroOperations(t *testing.T) {
	contract := parseContract(t)
	for _, operationID := range themeOperationIDs() {
		_ = findOperation(t, contract, operationID)
	}
}

// TestThemeContractDocumentsRetryAndConcurrencyHeaders verifies commands document replay safety.
func TestThemeContractDocumentsRetryAndConcurrencyHeaders(t *testing.T) {
	contract := parseContract(t)
	for _, operationID := range themeIdempotentOperationIDs() {
		operation := findOperation(t, contract, operationID)
		if !hasParameter(operation, "IdempotencyKeyHeader") {
			t.Fatalf("%s must document Idempotency-Key", operationID)
		}
	}
	for _, operationID := range themeVersionedOperationIDs() {
		operation := findOperation(t, contract, operationID)
		if !hasParameter(operation, "IfMatchHeader") {
			t.Fatalf("%s must document If-Match", operationID)
		}
	}
}

// TestThemeContractDeclaresRouteKinds verifies route-data contracts are enumerable.
func TestThemeContractDeclaresRouteKinds(t *testing.T) {
	schema := contractSchemas(t)["ThemeRouteKind"]
	enum, ok := schema["enum"].([]any)
	if !ok {
		t.Fatalf("ThemeRouteKind.enum is missing")
	}
	for _, route := range []string{
		"home",
		"forums.index",
		"forums.category",
		"forums.show",
		"threads.show",
		"threads.new",
		"tickets.index",
		"tickets.new",
		"tickets.show",
		"punishments.index",
		"punishments.show",
		"users.show",
		"search",
		"static.page",
		"not_found",
		"error",
		"maintenance",
		"auth.login",
		"auth.register",
		"auth.forgot_password",
		"auth.reset_password",
		"auth.verify_email",
		"auth.account_recovery",
	} {
		if !enumContains(enum, route) {
			t.Fatalf("ThemeRouteKind enum missing %q", route)
		}
	}
}

// TestThemeContractDeclaresAuthorizationScope verifies theme permissions can use theme scopes.
func TestThemeContractDeclaresAuthorizationScope(t *testing.T) {
	schema := contractSchemas(t)["AuthorizationObjectType"]
	enum, ok := schema["enum"].([]any)
	if !ok {
		t.Fatalf("AuthorizationObjectType.enum is missing")
	}
	if !enumContains(enum, "theme") {
		t.Fatalf("AuthorizationObjectType enum missing theme")
	}
}

// themeOperationIDs returns all theme operation IDs.
func themeOperationIDs() []string {
	return []string{
		"listThemes",
		"createTheme",
		"listInstalledThemes",
		"getTheme",
		"updateTheme",
		"deleteTheme",
		"installTheme",
		"disableTheme",
		"listThemeVersions",
		"createThemeVersionDraft",
		"importThemeVersion",
		"getThemeVersion",
		"validateThemeVersion",
		"archiveThemeVersion",
		"listThemeFiles",
		"createThemeFile",
		"getThemeFile",
		"updateThemeFile",
		"deleteThemeFile",
		"getThemeManifest",
		"getThemeFileByPath",
		"getThemeAssetByPath",
		"getThemeValidationReport",
		"getThemeDependencies",
		"getThemeIntegrity",
		"createThemePreviewToken",
		"getThemePreviewContext",
		"listThemeActivations",
		"activateThemeVersion",
		"rollbackThemeActivation",
		"getActiveThemeActivation",
		"getThemeRouteData",
		"listThemeSigningKeys",
		"createThemeSigningKey",
		"updateThemeSigningKey",
		"retireThemeSigningKey",
		"revokeThemeSigningKey",
	}
}

// themeIdempotentOperationIDs returns theme operations requiring replay protection.
func themeIdempotentOperationIDs() []string {
	return []string{
		"createTheme",
		"updateTheme",
		"deleteTheme",
		"installTheme",
		"disableTheme",
		"createThemeVersionDraft",
		"importThemeVersion",
		"validateThemeVersion",
		"archiveThemeVersion",
		"createThemeFile",
		"updateThemeFile",
		"deleteThemeFile",
		"createThemePreviewToken",
		"activateThemeVersion",
		"rollbackThemeActivation",
		"createThemeSigningKey",
		"updateThemeSigningKey",
		"retireThemeSigningKey",
		"revokeThemeSigningKey",
	}
}

// themeVersionedOperationIDs returns theme operations requiring optimistic concurrency.
func themeVersionedOperationIDs() []string {
	return []string{
		"updateTheme",
		"deleteTheme",
		"archiveThemeVersion",
		"updateThemeFile",
		"deleteThemeFile",
		"updateThemeSigningKey",
	}
}

// contractSchemas returns OpenAPI schemas as generic maps.
func contractSchemas(t *testing.T) map[string]map[string]any {
	t.Helper()
	var payload struct {
		Components struct {
			Schemas map[string]map[string]any `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(Document(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return payload.Components.Schemas
}

// enumContains reports whether enum contains expected.
func enumContains(enum []any, expected string) bool {
	return slices.ContainsFunc(enum, func(value any) bool {
		actual, _ := value.(string)
		return actual == expected
	})
}
