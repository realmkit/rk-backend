package signing

import (
	"testing"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// benchmarkManifestDigest stores manifest digest benchmark output.
var benchmarkManifestDigest domain.Digest

// benchmarkEnvelope stores decoded envelope benchmark output.
var benchmarkEnvelope Envelope

// BenchmarkManifestSHA256 measures canonical manifest JSON hashing.
func BenchmarkManifestSHA256(b *testing.B) {
	manifest := []byte(`{"version":"1.0.0","name":"Benchmark","settings":{"brand":"RealmKit","accent":"#33cc99"},"routes":["home","threads.show"]}`)

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		digest, err := ManifestSHA256(manifest)
		if err != nil {
			b.Fatalf("ManifestSHA256() error = %v", err)
		}
		benchmarkManifestDigest = digest
	}
}

// BenchmarkDecodeEnvelope measures detached signature envelope parsing.
func BenchmarkDecodeEnvelope(b *testing.B) {
	envelope := []byte(`{"algorithm":"ed25519","key_id":"realmkit:test","manifest_sha256":"abc123","signature":"signature","signed_at":"2026-01-01T00:00:00Z"}`)

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		decoded, err := DecodeEnvelope(envelope)
		if err != nil {
			b.Fatalf("DecodeEnvelope() error = %v", err)
		}
		benchmarkEnvelope = decoded
	}
}
