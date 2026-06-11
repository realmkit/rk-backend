package headers

import (
	"mime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Accept is the HTTP Accept header.
const Accept = "Accept"

// Authorization is the HTTP Authorization header.
const Authorization = "Authorization"

// ContentType is the HTTP Content-Type header.
const ContentType = "Content-Type"

// CorrelationID is the RealmKit correlation ID header.
const CorrelationID = "X-Correlation-Id"

// ETag is the HTTP ETag header.
const ETag = "ETag"

// IdempotencyKey is the idempotency key header.
const IdempotencyKey = "Idempotency-Key"

// IfMatch is the HTTP If-Match header.
const IfMatch = "If-Match"

// IfNoneMatch is the HTTP If-None-Match header.
const IfNoneMatch = "If-None-Match"

// Location is the HTTP Location header.
const Location = "Location"

// RequestID is the RealmKit request ID header.
const RequestID = "X-Request-Id"

// RetryAfter is the HTTP Retry-After header.
const RetryAfter = "Retry-After"

// RateLimitLimit is the rate limit size response header.
const RateLimitLimit = "RateLimit-Limit"

// RateLimitRemaining is the rate limit remaining response header.
const RateLimitRemaining = "RateLimit-Remaining"

// RateLimitReset is the rate limit reset response header.
const RateLimitReset = "RateLimit-Reset"

// RequestIDLocal is the Fiber local key for request IDs.
const RequestIDLocal = "realmkit.request_id"

// CorrelationIDLocal is the Fiber local key for correlation IDs.
const CorrelationIDLocal = "realmkit.correlation_id"

// Middleware preserves or creates request and correlation IDs.
func Middleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		requestID := normalizedHeader(ctx.Get(RequestID))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		correlationID := normalizedHeader(ctx.Get(CorrelationID))
		if correlationID == "" {
			correlationID = requestID
		}

		ctx.Locals(RequestIDLocal, requestID)
		ctx.Locals(CorrelationIDLocal, correlationID)
		ctx.Set(RequestID, requestID)
		ctx.Set(CorrelationID, correlationID)

		return ctx.Next()
	}
}

// CurrentRequestID returns the request ID for ctx.
func CurrentRequestID(ctx *fiber.Ctx) string {
	if value, ok := ctx.Locals(RequestIDLocal).(string); ok {
		return value
	}
	return ""
}

// CurrentCorrelationID returns the correlation ID for ctx.
func CurrentCorrelationID(ctx *fiber.Ctx) string {
	if value, ok := ctx.Locals(CorrelationIDLocal).(string); ok {
		return value
	}
	return ""
}

// RequireJSON enforces JSON Accept and Content-Type behavior.
func RequireJSON() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !acceptsJSON(ctx.Get(Accept)) {
			return fiber.NewError(fiber.StatusNotAcceptable, "Accept header must allow application/json.")
		}
		if hasRequestBody(ctx) && !isJSONMediaType(ctx.Get(ContentType)) {
			return fiber.NewError(fiber.StatusUnsupportedMediaType, "Content-Type must be application/json.")
		}

		return ctx.Next()
	}
}

// normalizedHeader trims header values.
func normalizedHeader(value string) string {
	return strings.TrimSpace(value)
}

// acceptsJSON reports whether value allows JSON responses.
func acceptsJSON(value string) bool {
	if strings.TrimSpace(value) == "" {
		return true
	}
	for _, part := range strings.Split(value, ",") {
		mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		if mediaType == "*/*" || mediaType == "application/*" || mediaType == "application/json" {
			return true
		}
	}
	return false
}

// hasRequestBody reports whether ctx carries a request body.
func hasRequestBody(ctx *fiber.Ctx) bool {
	return len(ctx.Body()) > 0
}

// isJSONMediaType reports whether value is application/json.
func isJSONMediaType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	return err == nil && mediaType == "application/json"
}
