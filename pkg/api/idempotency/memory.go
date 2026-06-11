package idempotency

import (
	"context"
	"sync"
	"time"
)

// MemoryStore stores idempotency records in process memory.
type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]Entry
}

// NewMemoryStore creates an in-memory idempotency store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: map[string]Entry{}}
}

// Reserve reserves key for fingerprint or returns an existing entry.
func (store *MemoryStore) Reserve(_ context.Context, key string, fingerprint string, ttl time.Duration) (Entry, bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if entry, ok := store.entries[key]; ok {
		if time.Now().Before(entry.ExpiresAt) {
			return entry, true, nil
		}
		delete(store.entries, key)
	}

	entry := Entry{
		Fingerprint: fingerprint,
		Complete:    false,
		ExpiresAt:   time.Now().Add(ttl),
	}
	store.entries[key] = entry
	return entry, false, nil
}

// Complete stores the response for key.
func (store *MemoryStore) Complete(_ context.Context, key string, entry Entry) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.entries[key] = entry
	return nil
}

// Ensure MemoryStore implements Store.
var _ Store = (*MemoryStore)(nil)
