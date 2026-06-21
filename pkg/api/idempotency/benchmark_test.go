package idempotency

import (
	"context"
	"strconv"
	"testing"
	"time"
)

// benchmarkEntry stores idempotency benchmark output.
var benchmarkEntry Entry

// benchmarkEntryExists stores idempotency reserve state.
var benchmarkEntryExists bool

// BenchmarkMemoryStoreReserveComplete measures reservation and completion over the in-memory store.
func BenchmarkMemoryStoreReserveComplete(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()
	entry := Entry{
		Fingerprint: "fingerprint",
		Status:      201,
		Body:        []byte(`{"ok":true}`),
		ContentType: "application/json",
		Complete:    true,
		ExpiresAt:   time.Now().Add(DefaultTTL),
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		key := "key-" + strconv.Itoa(index)
		reserved, exists, err := store.Reserve(ctx, key, entry.Fingerprint, DefaultTTL)
		if err != nil {
			b.Fatalf("Reserve() error = %v", err)
		}
		if exists {
			b.Fatalf("Reserve() exists = true for new key")
		}
		if err := store.Complete(ctx, key, entry); err != nil {
			b.Fatalf("Complete() error = %v", err)
		}
		benchmarkEntry = reserved
	}
}

// BenchmarkMemoryStoreReplay measures lookup of completed idempotency entries.
func BenchmarkMemoryStoreReplay(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()
	entry := Entry{
		Fingerprint: "fingerprint",
		Status:      200,
		Body:        []byte(`{"cached":true}`),
		ContentType: "application/json",
		Complete:    true,
		ExpiresAt:   time.Now().Add(DefaultTTL),
	}
	if err := store.Complete(ctx, "replay-key", entry); err != nil {
		b.Fatalf("Complete() error = %v", err)
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		got, exists, err := store.Reserve(ctx, "replay-key", entry.Fingerprint, DefaultTTL)
		if err != nil {
			b.Fatalf("Reserve() error = %v", err)
		}
		benchmarkEntry = got
		benchmarkEntryExists = exists
	}
}
