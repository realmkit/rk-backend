package openapi

import "testing"

// benchmarkOperationExists stores OpenAPI operation benchmark output.
var benchmarkOperationExists bool

// benchmarkDocumentBytes stores OpenAPI document benchmark output.
var benchmarkDocumentBytes []byte

// BenchmarkDocument measures defensive copying of the embedded OpenAPI document.
func BenchmarkDocument(b *testing.B) {
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkDocumentBytes = Document()
	}
}

// BenchmarkOperationExists measures contract parsing and Fiber path normalization lookup.
func BenchmarkOperationExists(b *testing.B) {
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		exists, err := OperationExists("GET", "/themes/:theme_id/versions/:version_id/manifest")
		if err != nil {
			b.Fatalf("OperationExists() error = %v", err)
		}
		benchmarkOperationExists = exists
	}
}
