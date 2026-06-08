package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"go.uber.org/zap"
)

// DefaultTTL is the default idempotency record lifetime.
const DefaultTTL = 24 * time.Hour

// MaxKeyLength is the maximum accepted idempotency key length.
const MaxKeyLength = 255

// MaxStoredBody is the largest response body cached for replay.
const MaxStoredBody = 1 << 20

// Entry stores one idempotency record.
type Entry struct {
	// Fingerprint is the request fingerprint.
	Fingerprint string `json:"fingerprint"`

	// Status is the stored response status.
	Status int `json:"status"`

	// Body is the stored response body.
	Body []byte `json:"body"`

	// ContentType is the stored response content type.
	ContentType string `json:"content_type"`

	// Complete reports whether the response can be replayed.
	Complete bool `json:"complete"`

	// ExpiresAt is the record expiry time.
	ExpiresAt time.Time `json:"expires_at"`
}

// Store reserves and completes idempotency records.
type Store interface {
	// Reserve reserves key for fingerprint or returns an existing entry.
	Reserve(ctx context.Context, key string, fingerprint string, ttl time.Duration) (Entry, bool, error)

	// Complete stores the response for key.
	Complete(ctx context.Context, key string, entry Entry) error
}

// Option changes idempotency middleware behavior.
type Option func(*settings)

// settings contains idempotency middleware settings.
type settings struct {
	log *zap.Logger
	ttl time.Duration
}

// MemoryStore stores idempotency records in process memory.
type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]Entry
	now     func() time.Time
}

// WithLogger configures structured idempotency logging.
func WithLogger(log *zap.Logger) Option {
	return func(settings *settings) {
		settings.log = log
	}
}

// WithTTL configures the idempotency record lifetime.
func WithTTL(ttl time.Duration) Option {
	return func(settings *settings) {
		settings.ttl = ttl
	}
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
func Middleware(store Store, options ...Option) fiber.Handler {
	settings := middlewareSettings(options)
	return func(ctx *fiber.Ctx) error {
		if !requiresIdempotencySupport(ctx.Method()) {
			return ctx.Next()
		}

		key, err := normalizedKey(ctx.Get(headers.IdempotencyKey))
		if err != nil {
			return problem.Write(ctx, problem.New(fiber.StatusBadRequest, "invalid_idempotency_key", err.Error()))
		}
		if key == "" {
			return ctx.Next()
		}

		fingerprint := fingerprintFor(ctx)
		entry, exists, err := store.Reserve(ctx.Context(), key, fingerprint, settings.ttl)
		if err != nil {
			return err
		}
		if exists {
			logDecision(settings.log, ctx, key, fingerprint, replayState(entry, fingerprint))
			return replayOrReject(ctx, entry, fingerprint)
		}

		logDecision(settings.log, ctx, key, fingerprint, "reserved")
		if err := ctx.Next(); err != nil {
			return err
		}

		if len(ctx.Response().Body()) > MaxStoredBody {
			return problem.Write(ctx, problem.New(fiber.StatusInternalServerError, "idempotency_response_too_large", "Response is too large to store for idempotent replay."))
		}

		return store.Complete(ctx.Context(), key, Entry{
			Fingerprint: fingerprint,
			Status:      ctx.Response().StatusCode(),
			Body:        append([]byte(nil), ctx.Response().Body()...),
			ContentType: string(ctx.Response().Header.ContentType()),
			Complete:    true,
			ExpiresAt:   time.Now().Add(settings.ttl),
		})
	}
}

// middlewareSettings applies idempotency middleware options.
func middlewareSettings(options []Option) settings {
	settings := settings{
		log: zap.NewNop(),
		ttl: DefaultTTL,
	}
	for _, option := range options {
		option(&settings)
	}
	if settings.log == nil {
		settings.log = zap.NewNop()
	}
	if settings.ttl <= 0 {
		settings.ttl = DefaultTTL
	}
	return settings
}

// requiresIdempotencySupport reports whether method can use idempotency keys.
func requiresIdempotencySupport(method string) bool {
	return method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch || method == fiber.MethodDelete
}

// normalizedKey validates and returns a trimmed idempotency key.
func normalizedKey(value string) (string, error) {
	key := strings.TrimSpace(value)
	if key == "" {
		return "", nil
	}
	if len(key) > MaxKeyLength {
		return "", fmt.Errorf("Idempotency-Key exceeds maximum length")
	}
	return key, nil
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

// replayState returns the replay state for logging.
func replayState(entry Entry, fingerprint string) string {
	if entry.Fingerprint != fingerprint {
		return "conflict"
	}
	if !entry.Complete {
		return "in_progress"
	}
	return "replay"
}

// keyHash returns a safe idempotency key hash for logs.
func keyHash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// logDecision writes an idempotency decision without exposing raw keys.
func logDecision(log *zap.Logger, ctx *fiber.Ctx, key string, fingerprint string, state string) {
	log.Info("idempotency decision",
		zap.String("component", "idempotency"),
		zap.String("key_hash", keyHash(key)),
		zap.String("fingerprint", fingerprint),
		zap.String("state", state),
		zap.String("method", ctx.Method()),
		zap.String("path", ctx.Path()),
		zap.String("request_id", headers.CurrentRequestID(ctx)),
		zap.String("correlation_id", headers.CurrentCorrelationID(ctx)),
	)
}
