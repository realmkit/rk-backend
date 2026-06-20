package signing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// Envelope is the detached package signature document.
type Envelope struct {
	Algorithm      string `json:"algorithm"`
	KeyID          string `json:"key_id"`
	ManifestSHA256 string `json:"manifest_sha256"`
	Signature      string `json:"signature"`
	SignedAt       string `json:"signed_at"`
}

// CanonicalManifestJSON returns deterministic JSON used for hashing and signing.
func CanonicalManifestJSON(value []byte) ([]byte, error) {
	var document any
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("parse manifest json: %w", err)
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("canonicalize manifest json: %w", err)
	}
	return encoded, nil
}

// ManifestSHA256 returns a lowercase hex SHA-256 digest for canonical manifest JSON.
func ManifestSHA256(value []byte) (domain.Digest, error) {
	canonical, err := CanonicalManifestJSON(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(canonical)
	return domain.Digest(hex.EncodeToString(hash[:])), nil
}

// DecodeEnvelope parses a package signature envelope.
func DecodeEnvelope(value []byte) (Envelope, error) {
	var envelope Envelope
	if err := json.Unmarshal(value, &envelope); err != nil {
		return Envelope{}, fmt.Errorf("parse signature envelope: %w", err)
	}
	if strings.TrimSpace(envelope.KeyID) == "" || strings.TrimSpace(envelope.Signature) == "" {
		return Envelope{}, fmt.Errorf("signature envelope key_id and signature are required")
	}
	return envelope, nil
}

// NormalizeDigest returns a lowercase digest without optional prefixes.
func NormalizeDigest(value string) domain.Digest {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.TrimPrefix(normalized, "sha256:")
	normalized = strings.TrimPrefix(normalized, "sha256-")
	return domain.Digest(normalized)
}
