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
	ok, err := OperationExists("GET", "/health")
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("OperationExists() = false, want true")
	}
}

// TestOperationExistsNormalizesFiberParameters verifies Fiber paths match OpenAPI paths.
func TestOperationExistsNormalizesFiberParameters(t *testing.T) {
	ok, err := OperationExists("GET", "/metadata/metafield-definitions/:definition_id")
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("OperationExists() = false, want true")
	}
}

// TestPunishmentOperationsExist verifies punishment routes are documented.
func TestPunishmentOperationsExist(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{method: "POST", path: "/punishment-definitions"},
		{method: "GET", path: "/punishment-definitions"},
		{method: "GET", path: "/punishment-definitions/:definition_id"},
		{method: "PATCH", path: "/punishment-definitions/:definition_id"},
		{method: "DELETE", path: "/punishment-definitions/:definition_id"},
		{method: "POST", path: "/punishment-definitions/:definition_id/actions/reorder"},
		{method: "POST", path: "/punishments"},
		{method: "GET", path: "/punishments"},
		{method: "GET", path: "/punishments/:punishment_id"},
		{method: "PATCH", path: "/punishments/:punishment_id"},
		{method: "POST", path: "/punishments/:punishment_id/revoke"},
		{method: "GET", path: "/users/:user_id/punishments"},
		{method: "GET", path: "/users/:user_id/punishments/active"},
		{method: "POST", path: "/punishments/restrictions/check"},
		{method: "GET", path: "/users/:user_id/punishments/restrictions"},
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

// TestPunishmentOperationsUseConcernTags verifies punishment routes are grouped.
func TestPunishmentOperationsUseConcernTags(t *testing.T) {
	expected := map[string]string{
		"checkPunishmentRestriction":         "punishment-restrictions",
		"createPunishmentDefinition":         "punishment-definitions",
		"deletePunishmentDefinition":         "punishment-definitions",
		"getPunishment":                      "punishments",
		"getPunishmentDefinition":            "punishment-definitions",
		"issuePunishment":                    "punishments",
		"listPunishmentDefinitions":          "punishment-definitions",
		"listPunishments":                    "punishments",
		"listUserActivePunishments":          "punishments",
		"listUserPunishmentRestrictions":     "punishment-restrictions",
		"listUserPunishments":                "punishments",
		"reorderPunishmentDefinitionActions": "punishment-definitions",
		"revokePunishment":                   "punishments",
		"updatePunishment":                   "punishments",
		"updatePunishmentDefinition":         "punishment-definitions",
	}
	requiredTags := map[string]bool{
		"punishment-definitions":  false,
		"punishment-restrictions": false,
		"punishments":             false,
	}

	var contract struct {
		Tags []struct {
			Name string `json:"name"`
		} `json:"tags"`
		Paths map[string]map[string]struct {
			OperationID string   `json:"operationId"`
			Tags        []string `json:"tags"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(Document(), &contract); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	for _, tag := range contract.Tags {
		if _, ok := requiredTags[tag.Name]; ok {
			requiredTags[tag.Name] = true
		}
	}
	for tag, found := range requiredTags {
		if !found {
			t.Fatalf("punishment tag %q is not declared", tag)
		}
	}

	seen := map[string]bool{}
	for _, methods := range contract.Paths {
		for _, operation := range methods {
			want, ok := expected[operation.OperationID]
			if !ok {
				continue
			}
			seen[operation.OperationID] = true
			if len(operation.Tags) != 1 || operation.Tags[0] != want {
				t.Fatalf("%s tags = %v, want [%s]", operation.OperationID, operation.Tags, want)
			}
		}
	}
	for operationID := range expected {
		if !seen[operationID] {
			t.Fatalf("punishment operation %q was not found", operationID)
		}
	}
}

// TestForumMilestoneOneOperationsExist verifies forum structure routes are documented.
func TestForumMilestoneOneOperationsExist(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{method: "GET", path: "/forums/tree"},
		{method: "POST", path: "/forum-categories"},
		{method: "POST", path: "/forum-categories/reorder"},
		{method: "PATCH", path: "/forum-categories/:category_id"},
		{method: "POST", path: "/forums"},
		{method: "POST", path: "/forums/reorder"},
		{method: "POST", path: "/forums/:forum_id/move"},
		{method: "GET", path: "/forums/:forum_id/threads"},
		{method: "POST", path: "/forums/:forum_id/threads"},
		{method: "PATCH", path: "/threads/:thread_id"},
		{method: "GET", path: "/threads/:thread_id/posts"},
		{method: "POST", path: "/threads/:thread_id/posts"},
		{method: "POST", path: "/threads/:thread_id/read"},
		{method: "PATCH", path: "/posts/:post_id"},
		{method: "PUT", path: "/posts/:post_id/like"},
		{method: "DELETE", path: "/posts/:post_id/like"},
		{method: "GET", path: "/posts/:post_id/revisions"},
		{method: "GET", path: "/forums/latest-posts"},
		{method: "GET", path: "/forums/search"},
		{method: "GET", path: "/forums/unread-summary"},
		{method: "POST", path: "/forums/:forum_id/read"},
		{method: "GET", path: "/forums/:forum_id/settings"},
		{method: "PATCH", path: "/forums/:forum_id/settings"},
		{method: "GET", path: "/forums/:forum_id/permissions"},
		{method: "PUT", path: "/forums/:forum_id/permissions"},
		{method: "POST", path: "/forums/:forum_id/permissions/simulate"},
		{method: "GET", path: "/forums/:forum_id/latest-posts"},
		{method: "GET", path: "/forums/:forum_id/posts/most-liked"},
		{method: "GET", path: "/forums/:forum_id/search"},
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

// TestForumOperationsUseConcernTags verifies forum routes stay grouped by concern.
func TestForumOperationsUseConcernTags(t *testing.T) {
	expected := map[string]string{
		"createForum":             "forums",
		"createForumCategory":     "forum-categories",
		"createForumThread":       "forum-threads",
		"createThreadReply":       "forum-posts",
		"deleteForum":             "forums",
		"deleteForumCategory":     "forum-categories",
		"deletePost":              "forum-posts",
		"deleteThread":            "forum-threads",
		"getForum":                "forums",
		"getForumCategory":        "forum-categories",
		"getForumPermissions":     "forum-admin",
		"getForumSettings":        "forum-admin",
		"getForumTree":            "forums",
		"getForumUnreadSummary":   "forum-interactions",
		"getPost":                 "forum-posts",
		"getThread":               "forum-threads",
		"likePost":                "forum-interactions",
		"listForumCategories":     "forum-categories",
		"listForumLatestPosts":    "forum-interactions",
		"listForumThreads":        "forum-threads",
		"listForums":              "forums",
		"listLatestPosts":         "forum-interactions",
		"listMostLikedPosts":      "forum-interactions",
		"listPostRevisions":       "forum-posts",
		"listThreadPosts":         "forum-posts",
		"markForumRead":           "forum-interactions",
		"markThreadRead":          "forum-interactions",
		"moveForum":               "forums",
		"reorderForumCategories":  "forum-categories",
		"reorderForums":           "forums",
		"searchForum":             "forum-search",
		"searchForums":            "forum-search",
		"simulateForumPermission": "forum-admin",
		"unlikePost":              "forum-interactions",
		"updateForum":             "forums",
		"updateForumCategory":     "forum-categories",
		"updateForumPermissions":  "forum-admin",
		"updateForumSettings":     "forum-admin",
		"updatePost":              "forum-posts",
		"updateThread":            "forum-threads",
	}
	requiredTags := map[string]bool{
		"forum-admin":        false,
		"forum-categories":   false,
		"forum-interactions": false,
		"forum-posts":        false,
		"forum-search":       false,
		"forum-threads":      false,
		"forums":             false,
	}

	var contract struct {
		Tags []struct {
			Name string `json:"name"`
		} `json:"tags"`
		Paths map[string]map[string]struct {
			OperationID string   `json:"operationId"`
			Tags        []string `json:"tags"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(Document(), &contract); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	for _, tag := range contract.Tags {
		if _, ok := requiredTags[tag.Name]; ok {
			requiredTags[tag.Name] = true
		}
	}
	for tag, found := range requiredTags {
		if !found {
			t.Fatalf("forum tag %q is not declared", tag)
		}
	}

	seen := map[string]bool{}
	for _, methods := range contract.Paths {
		for _, operation := range methods {
			want, ok := expected[operation.OperationID]
			if !ok {
				continue
			}
			seen[operation.OperationID] = true
			if len(operation.Tags) != 1 || operation.Tags[0] != want {
				t.Fatalf(
					"%s tags = %v, want [%s]",
					operation.OperationID,
					operation.Tags,
					want,
				)
			}
		}
	}
	for operationID := range expected {
		if !seen[operationID] {
			t.Fatalf("forum operation %q was not found", operationID)
		}
	}
}
