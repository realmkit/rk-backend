package openapi

import (
	_ "embed"
	"encoding/json"
	"strings"
)

// DocumentFile is the embedded OpenAPI document path.
const DocumentFile = "realmkit.v1.json"

// document contains the RealmKit v1 OpenAPI contract.
//
//go:embed realmkit.v1.json
var document []byte

// Document returns the embedded OpenAPI document bytes.
func Document() []byte {
	copyDocument := make([]byte, len(document))
	copy(copyDocument, document)
	return copyDocument
}

// OperationExists reports whether method and path exist in the contract.
func OperationExists(method string, path string) (bool, error) {
	var contract contractDocument
	if err := json.Unmarshal(document, &contract); err != nil {
		return false, err
	}

	operations, ok := contract.Paths[path]
	if !ok {
		operations, ok = contract.Paths[normalizeFiberPath(path)]
	}
	if !ok {
		return false, nil
	}
	_, ok = operations[strings.ToLower(method)]
	return ok, nil
}

// normalizeFiberPath converts Fiber route parameters to OpenAPI parameters.
func normalizeFiberPath(path string) string {
	parts := strings.Split(path, "/")
	for index, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[index] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

// contractDocument contains the OpenAPI fields needed by contract checks.
type contractDocument struct {
	Paths map[string]map[string]any `json:"paths"` // Paths stores the paths value.
}
