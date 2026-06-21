package operations

import (
	"testing"

	"github.com/google/uuid"
)

// benchmarkViewIncrements stores parsed thread view increments.
var benchmarkViewIncrements map[uuid.UUID]int64

// benchmarkViewTotal stores parsed thread view totals.
var benchmarkViewTotal int64

// BenchmarkThreadViewIncrements measures Redis-drained counter parsing and invalid-key filtering.
func BenchmarkThreadViewIncrements(b *testing.B) {
	raw := make(map[string]int64, 130)
	for index := 0; index < 128; index++ {
		raw[uuid.NewString()] = int64(index + 1)
	}
	raw["bad-id"] = 50
	raw[uuid.NewString()] = -1

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		increments, total := threadViewIncrements(raw)
		benchmarkViewIncrements = increments
		benchmarkViewTotal = total
	}
}
