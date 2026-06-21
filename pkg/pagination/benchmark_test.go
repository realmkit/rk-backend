package pagination

import "testing"

// benchmarkPage stores pagination benchmark output.
var benchmarkPage Page

// BenchmarkNew measures pagination normalization.
func BenchmarkNew(b *testing.B) {
	request := Request{Limit: 250, Cursor: "  opaque-cursor  "}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		page, err := New(request)
		if err != nil {
			b.Fatalf("New() error = %v", err)
		}
		benchmarkPage = page
	}
}
