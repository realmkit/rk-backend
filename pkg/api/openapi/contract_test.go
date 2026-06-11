package openapi

import (
	"encoding/json"
	"strings"
	"testing"
)

type operationContract struct {
	OperationID string           `json:"operationId"`
	Tags        []string         `json:"tags"`
	Parameters  []map[string]any `json:"parameters"`
	Description string           `json:"description"`
	Responses   map[string]any   `json:"responses"`
}

type apiContract struct {
	Servers []struct {
		URL string `json:"url"`
	} `json:"servers"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`
	Paths map[string]map[string]operationContract `json:"paths"`
}

// TestContractUsesGatewayVersioning verifies service routes remain unversioned.
func TestContractUsesGatewayVersioning(t *testing.T) {
	contract := parseContract(t)
	if len(contract.Servers) == 0 || contract.Servers[0].URL != "/" {
		t.Fatalf("servers = %+v, want root server URL for gateway-owned versioning", contract.Servers)
	}
	for path := range contract.Paths {
		if strings.HasPrefix(path, "/api/") {
			t.Fatalf("path %q must not include service-owned API versioning", path)
		}
	}
}

// TestContractOperationsHaveDeclaredTagsAndProblemResponses verifies route grouping and baseline errors.
func TestContractOperationsHaveDeclaredTagsAndProblemResponses(t *testing.T) {
	contract := parseContract(t)
	declared := map[string]bool{}
	for _, tag := range contract.Tags {
		declared[tag.Name] = true
	}
	for path, methods := range contract.Paths {
		for method, operation := range methods {
			if len(operation.Tags) != 1 {
				t.Fatalf("%s %s tags = %v, want exactly one concern tag", method, path, operation.Tags)
			}
			if !declared[operation.Tags[0]] {
				t.Fatalf("%s %s tag %q is not declared", method, path, operation.Tags[0])
			}
			if operation.OperationID == "" {
				t.Fatalf("%s %s missing operationId", method, path)
			}
			for _, code := range []string{"429", "500"} {
				if _, ok := operation.Responses[code]; !ok {
					t.Fatalf("%s %s missing %s response", method, path, code)
				}
			}
		}
	}
}

// TestContractDocumentsOperationalHeaders verifies high-risk commands document replay and version headers.
func TestContractDocumentsOperationalHeaders(t *testing.T) {
	contract := parseContract(t)
	idempotent := []string{
		"createAssetUploadIntent",
		"completeAssetUpload",
		"createForumThread",
		"createThreadReply",
		"issuePunishment",
		"revokePunishment",
		"createTicket",
		"createTicketMessage",
		"closeTicket",
	}
	versioned := []string{
		"updateAsset",
		"deleteAsset",
		"updateForum",
		"deleteThread",
		"updatePunishment",
		"revokePunishment",
		"updateTicketDefinition",
		"closeTicket",
		"pauseCronJob",
	}
	for _, operationID := range idempotent {
		operation := findOperation(t, contract, operationID)
		if !hasParameter(operation, "IdempotencyKeyHeader") || !strings.Contains(operation.Description, "Idempotency-Key") {
			t.Fatalf("%s must document Idempotency-Key behavior", operationID)
		}
	}
	for _, operationID := range versioned {
		operation := findOperation(t, contract, operationID)
		if !hasParameter(operation, "IfMatchHeader") || !strings.Contains(operation.Description, "If-Match") {
			t.Fatalf("%s must document If-Match behavior", operationID)
		}
	}
}

func parseContract(t *testing.T) apiContract {
	t.Helper()
	var contract apiContract
	if err := json.Unmarshal(Document(), &contract); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return contract
}

func findOperation(t *testing.T, contract apiContract, operationID string) operationContract {
	t.Helper()
	for _, methods := range contract.Paths {
		for _, operation := range methods {
			if operation.OperationID == operationID {
				return operation
			}
		}
	}
	t.Fatalf("operation %q was not found", operationID)
	return operationContract{}
}

func hasParameter(operation operationContract, component string) bool {
	for _, parameter := range operation.Parameters {
		ref, _ := parameter["$ref"].(string)
		if strings.HasSuffix(ref, "/"+component) {
			return true
		}
	}
	return false
}
