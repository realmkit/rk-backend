package storage

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Object describes an object to write to storage.
type Object struct {
	// Key is the S3 object key.
	Key string

	// ContentType is the object media type.
	ContentType string

	// SizeBytes is the expected object size.
	SizeBytes int64

	// Metadata contains provider object metadata.
	Metadata map[string]string
}

// StoredObject describes a stored object.
type StoredObject struct {
	// Key is the S3 object key.
	Key string

	// ETag is the provider entity tag.
	ETag string
}

// PresignPutRequest requests a presigned upload URL.
type PresignPutRequest struct {
	// Key is the S3 object key.
	Key string

	// ContentType is the required upload media type.
	ContentType string

	// ExpiresIn is the signed URL lifetime.
	ExpiresIn time.Duration
}

// PresignedRequest contains a presigned HTTP request.
type PresignedRequest struct {
	// Method is the HTTP method.
	Method string

	// URL is the presigned URL.
	URL string

	// Headers contains required request headers.
	Headers map[string]string

	// ExpiresAt is when the request expires.
	ExpiresAt time.Time
}

// ObjectInfo describes an object stored in S3.
type ObjectInfo struct {
	// Key is the S3 object key.
	Key string

	// ETag is the provider entity tag.
	ETag string

	// ContentType is the object media type.
	ContentType string

	// SizeBytes is the object size.
	SizeBytes int64

	// Metadata contains provider object metadata.
	Metadata map[string]string
}

// Store reads, writes, deletes, and signs S3 objects.
type Store interface {
	// Health verifies the storage backend is reachable.
	Health(ctx context.Context) error

	// Put stores object bytes.
	Put(ctx context.Context, object Object, body io.Reader) (StoredObject, error)

	// Delete deletes an object by key.
	Delete(ctx context.Context, key string) error

	// PresignPut creates a presigned upload request.
	PresignPut(ctx context.Context, request PresignPutRequest) (PresignedRequest, error)

	// PresignGet creates a presigned download URL.
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)

	// Head returns object metadata.
	Head(ctx context.Context, key string) (ObjectInfo, error)
}

// putHeaders returns headers for a presigned upload.
func putHeaders(contentType string) map[string]string {
	if contentType == "" {
		return nil
	}
	return map[string]string{http.CanonicalHeaderKey("content-type"): contentType}
}
