package domain

import (
	"fmt"
	"testing"
)

// BenchmarkCalculateVersionIntegritySHA256 measures deterministic version hashing.
func BenchmarkCalculateVersionIntegritySHA256(b *testing.B) {
	files := make([]IntegrityFile, 0, 256)
	for index := 0; index < cap(files); index++ {
		files = append(files, IntegrityFile{
			Path:          FilePath(fmt.Sprintf("templates/page_%03d.liquid", index)),
			ContentSHA256: Digest(fmt.Sprintf("%064x", index)),
		})
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if digest := CalculateVersionIntegritySHA256(files); digest == "" {
			b.Fatalf("CalculateVersionIntegritySHA256() returned empty digest")
		}
	}
}
