package openapi

import (
	_ "embed"
	"encoding/json"
	"strings"
)

// DocumentFile is the embedded OpenAPI document path.
const DocumentFile = "gamehub.v1.json"

// document contains the GameHub v1 OpenAPI contract.
//
//go:embed gamehub.v1.json
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
		return false, nil
	}
	_, ok = operations[strings.ToLower(method)]
	return ok, nil
}

// contractDocument contains the OpenAPI fields needed by contract checks.
type contractDocument struct {
	Paths map[string]map[string]any `json:"paths"`
}
