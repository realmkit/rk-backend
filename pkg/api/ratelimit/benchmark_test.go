package ratelimit

import (
	"context"
	"strconv"
	"testing"
	"time"
)

// benchmarkRateLimitDecision stores rate-limit benchmark output.
var benchmarkRateLimitDecision Decision

// BenchmarkMemoryStoreAllowSingleKey measures hot-key in-memory rate limit decisions.
func BenchmarkMemoryStoreAllowSingleKey(b *testing.B) {
	store := NewMemoryStore()
	store.now = func() time.Time { return time.Unix(100, 0).UTC() }
	policy := Policy{Limit: b.N + 1, Window: time.Minute}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		decision, err := store.Allow(ctx, "user:1", policy)
		if err != nil {
			b.Fatalf("Allow() error = %v", err)
		}
		benchmarkRateLimitDecision = decision
	}
}

// BenchmarkMemoryStoreAllowManyKeys measures rate-limit window creation for many subjects.
func BenchmarkMemoryStoreAllowManyKeys(b *testing.B) {
	store := NewMemoryStore()
	store.now = func() time.Time { return time.Unix(100, 0).UTC() }
	policy := Policy{Limit: 100, Window: time.Minute}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		decision, err := store.Allow(ctx, "user:"+strconv.Itoa(index), policy)
		if err != nil {
			b.Fatalf("Allow() error = %v", err)
		}
		benchmarkRateLimitDecision = decision
	}
}
