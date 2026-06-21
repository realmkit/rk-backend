package delivery

import (
	"context"
	"testing"
)

// benchmarkManifestResult stores delivery manifest benchmark output.
var benchmarkManifestResult ManifestResult

// BenchmarkManifest measures render manifest assembly through the delivery service.
func BenchmarkManifest(b *testing.B) {
	repositories, themeID, versionID := deliveryRepositories()
	service := NewService(repositories, fixedClock())
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		manifest, err := service.Manifest(ctx, themeID, versionID)
		if err != nil {
			b.Fatalf("Manifest() error = %v", err)
		}
		benchmarkManifestResult = manifest
	}
}
