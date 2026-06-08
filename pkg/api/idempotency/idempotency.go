package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
)

// DefaultTTL is the default idempotency record lifetime.
const DefaultTTL = 24 * time.Hour

// Entry stores one idempotency record.
type Entry struct {
	// Fingerprint is the request fingerprint.
	Fingerprint string

	// Status is the stored response status.
	Status int

	// Body is the stored response body.
	Body []byte

	// ContentType is the stored response content type.
	ContentType string

	// Complete reports whether the response can be replayed.
	Complete bool

	// ExpiresAt is the record expiry time.
	ExpiresAt time.Time
}

// Store reserves and completes idempotency records.
type Store interface {
	// Reserve reserves key for fingerprint or returns an existing entry.
	Reserve(ctx context.Context, key string, fingerprint string, ttl time.Duration) (Entry, bool, error)

	// Complete stores the response for key.
	Complete(ctx context.Context, key string, entry Entry) error
}

// MemoryStore stores idempotency records in process memory.
type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]Entry
	now     func() time.Time
}

// NewMemoryStore creates an in-memory idempotency store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Reserve reserves key for fingerprint or returns an existing entry.
func (store *MemoryStore) Reserve(_ context.Context, key string, fingerprint string, ttl time.Duration) (Entry, bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if entry, ok := store.entries[key]; ok && entry.ExpiresAt.After(store.now()) {
		return entry, true, nil
	}

	entry := Entry{Fingerprint: fingerprint, ExpiresAt: store.now().Add(ttl)}
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

// Middleware returns idempotency middleware.
func Middleware(store Store) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !requiresIdempotencySupport(ctx.Method()) {
			return ctx.Next()
		}

		key := ctx.Get(headers.IdempotencyKey)
		if key == "" {
			return ctx.Next()
		}

		fingerprint := fingerprintFor(ctx)
		entry, exists, err := store.Reserve(ctx.Context(), key, fingerprint, DefaultTTL)
		if err != nil {
			return err
		}
		if exists {
			return replayOrReject(ctx, entry, fingerprint)
		}

		if err := ctx.Next(); err != nil {
			return err
		}

		return store.Complete(ctx.Context(), key, Entry{
			Fingerprint: fingerprint,
			Status:      ctx.Response().StatusCode(),
			Body:        append([]byte(nil), ctx.Response().Body()...),
			ContentType: string(ctx.Response().Header.ContentType()),
			Complete:    true,
			ExpiresAt:   time.Now().Add(DefaultTTL),
		})
	}
}

// requiresIdempotencySupport reports whether method can use idempotency keys.
func requiresIdempotencySupport(method string) bool {
	return method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch || method == fiber.MethodDelete
}

// fingerprintFor returns the request fingerprint.
func fingerprintFor(ctx *fiber.Ctx) string {
	sum := sha256.Sum256(append([]byte(ctx.Method()+" "+ctx.Path()+" "), ctx.Body()...))
	return hex.EncodeToString(sum[:])
}

// replayOrReject handles existing idempotency records.
func replayOrReject(ctx *fiber.Ctx, entry Entry, fingerprint string) error {
	if entry.Fingerprint != fingerprint {
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "idempotency_conflict", "Idempotency key was reused with a different request."))
	}
	if !entry.Complete {
		return problem.Write(ctx, problem.New(fiber.StatusConflict, "idempotency_in_progress", "Idempotent request is still in progress."))
	}
	if entry.ContentType != "" {
		ctx.Set(headers.ContentType, entry.ContentType)
	}
	return ctx.Status(entry.Status).Send(entry.Body)
}
