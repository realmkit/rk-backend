package openapi

import (
	"encoding/json"
	"testing"
)

// TestSearchableListOperationsExposeFilters verifies search indicators.
func TestSearchableListOperationsExposeFilters(t *testing.T) {
	contract := openAPITestDocument(t)
	tests := []struct {
		path  string
		names []string
	}{
		{path: "/groups", names: []string{"q", "has_icon", "min_weight", "max_weight", "sort", "direction"}},
		{path: "/assets", names: []string{"q", "visibility", "sort", "direction"}},
		{path: "/users", names: []string{"q", "status", "sort", "direction"}},
		{path: "/punishment-definitions", names: []string{"q", "sort", "direction"}},
		{path: "/punishments", names: []string{"q", "target_user_id", "sort", "direction"}},
		{path: "/ticket-definitions", names: []string{"q", "kind", "status", "sort", "direction"}},
		{path: "/tickets", names: []string{"q", "submitter_user_id", "assignee_user_id", "sort", "direction"}},
		{path: "/forums/search", names: []string{"q", "query"}},
		{path: "/forums/{forum_id}/search", names: []string{"q", "query"}},
	}
	for _, test := range tests {
		parameters := operationParameterNames(t, contract, test.path)
		for _, name := range test.names {
			if !parameters[name] {
				t.Fatalf("%s missing query/header/path parameter %q in %v", test.path, name, parameters)
			}
		}
	}
}

// TestSearchableListSchemasExposeAppliedSearch verifies response indicators.
func TestSearchableListSchemasExposeAppliedSearch(t *testing.T) {
	contract := openAPITestDocument(t)
	for _, schemaName := range []string{
		"AssetList",
		"GroupList",
		"PunishmentDefinitionList",
		"PunishmentList",
		"TicketDefinitionList",
		"TicketList",
		"UserList",
	} {
		schema := contract.Components.Schemas[schemaName]
		for _, property := range []string{"query", "sort", "direction", "next_page_token"} {
			if _, ok := schema.Properties[property]; !ok {
				t.Fatalf("%s missing property %q", schemaName, property)
			}
		}
	}
}

// openAPITestDocument decodes the OpenAPI document.
func openAPITestDocument(t *testing.T) openAPITestContract {
	t.Helper()
	var contract openAPITestContract
	if err := json.Unmarshal(Document(), &contract); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return contract
}

// operationParameterNames returns declared parameter names for one operation.
func operationParameterNames(t *testing.T, contract openAPITestContract, path string) map[string]bool {
	t.Helper()
	operation, ok := contract.Paths[path]["get"]
	if !ok {
		t.Fatalf("GET %s is not documented", path)
	}
	names := map[string]bool{}
	for _, parameter := range operation.Parameters {
		if parameter.Ref != "" {
			parameter = dereferenceParameter(t, contract, parameter.Ref)
		}
		names[parameter.Name] = true
	}
	return names
}

// dereferenceParameter resolves a local OpenAPI parameter reference.
func dereferenceParameter(t *testing.T, contract openAPITestContract, reference string) openAPITestParameter {
	t.Helper()
	const prefix = "#/components/parameters/"
	if len(reference) <= len(prefix) || reference[:len(prefix)] != prefix {
		t.Fatalf("unsupported parameter reference %q", reference)
	}
	parameter, ok := contract.Components.Parameters[reference[len(prefix):]]
	if !ok {
		t.Fatalf("parameter reference %q not found", reference)
	}
	return parameter
}

// openAPITestContract contains the fields needed by filtering tests.
type openAPITestContract struct {
	Paths      map[string]map[string]openAPITestOperation `json:"paths"`
	Components struct {
		Parameters map[string]openAPITestParameter `json:"parameters"`
		Schemas    map[string]openAPITestSchema    `json:"schemas"`
	} `json:"components"`
}

// openAPITestOperation contains one operation subset.
type openAPITestOperation struct {
	Parameters []openAPITestParameter `json:"parameters"`
}

// openAPITestParameter contains one parameter subset.
type openAPITestParameter struct {
	Ref  string `json:"$ref"`
	Name string `json:"name"`
}

// openAPITestSchema contains one schema subset.
type openAPITestSchema struct {
	Properties map[string]any `json:"properties"`
}
