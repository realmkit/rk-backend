package application

import (
	"testing"

	"github.com/realmkit/rk-backend/module/themes/domain"
)

// benchmarkSigningKeys stores parsed signing key benchmark output.
var benchmarkSigningKeys []domain.ThemeSigningKey

// BenchmarkConfigSigningKeys measures environment JSON parsing into domain signing keys.
func BenchmarkConfigSigningKeys(b *testing.B) {
	cfg := Config{SigningKeysJSON: `[` +
		`{"key_id":"realmkit:primary","public_key":"primary-public-key","trust_level":"operator","status":"trusted","not_before":"2026-01-01T00:00:00Z"},` +
		`{"key_id":"realmkit:secondary","public_key":"secondary-public-key","trust_level":"community","status":"retired","not_after":"2027-01-01T00:00:00Z"}` +
		`]`}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		keys, err := cfg.SigningKeys()
		if err != nil {
			b.Fatalf("SigningKeys() error = %v", err)
		}
		benchmarkSigningKeys = keys
	}
}
