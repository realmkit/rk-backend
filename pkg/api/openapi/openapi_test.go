package openapi

import (
	"encoding/json"
	"testing"
)

// TestDocumentReturnsCopy verifies callers cannot mutate the embedded document.
func TestDocumentReturnsCopy(t *testing.T) {
	first := Document()
	second := Document()
	first[0] = 'x'

	if second[0] == 'x' {
		t.Fatalf("Document() returned shared backing storage")
	}
}

// TestDocumentIsOpenAPI31 verifies the embedded document is valid JSON and OpenAPI 3.1.
func TestDocumentIsOpenAPI31(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal(Document(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload["openapi"] != "3.1.0" {
		t.Fatalf("openapi = %v, want %v", payload["openapi"], "3.1.0")
	}
}

// TestOperationExistsFindsDocumentedOperations verifies operation lookup.
func TestOperationExistsFindsDocumentedOperations(t *testing.T) {
	ok, err := OperationExists("GET", "/api/v1/health")
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("OperationExists() = false, want true")
	}
}

// TestOperationExistsNormalizesFiberParameters verifies Fiber paths match OpenAPI paths.
func TestOperationExistsNormalizesFiberParameters(t *testing.T) {
	ok, err := OperationExists("GET", "/api/v1/metadata/metafield-definitions/:definition_id")
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("OperationExists() = false, want true")
	}
}

// TestForumMilestoneOneOperationsExist verifies forum structure routes are documented.
func TestForumMilestoneOneOperationsExist(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{method: "GET", path: "/api/v1/forums/tree"},
		{method: "POST", path: "/api/v1/forum-categories"},
		{method: "POST", path: "/api/v1/forum-categories/reorder"},
		{method: "PATCH", path: "/api/v1/forum-categories/:category_id"},
		{method: "POST", path: "/api/v1/forums"},
		{method: "POST", path: "/api/v1/forums/reorder"},
		{method: "POST", path: "/api/v1/forums/:forum_id/move"},
		{method: "GET", path: "/api/v1/forums/:forum_id/threads"},
		{method: "POST", path: "/api/v1/forums/:forum_id/threads"},
		{method: "PATCH", path: "/api/v1/threads/:thread_id"},
		{method: "GET", path: "/api/v1/threads/:thread_id/posts"},
		{method: "POST", path: "/api/v1/threads/:thread_id/posts"},
		{method: "PATCH", path: "/api/v1/posts/:post_id"},
		{method: "GET", path: "/api/v1/posts/:post_id/revisions"},
	}

	for _, tt := range tests {
		ok, err := OperationExists(tt.method, tt.path)
		if err != nil {
			t.Fatalf("OperationExists(%q, %q) error = %v", tt.method, tt.path, err)
		}
		if !ok {
			t.Fatalf("OperationExists(%q, %q) = false, want true", tt.method, tt.path)
		}
	}
}
