package storage

import (
	"context"
	"testing"
)

// TestNewS3StoreRequiresMandatoryConfig verifies S3 config guardrails.
func TestNewS3StoreRequiresMandatoryConfig(t *testing.T) {
	if _, err := NewS3Store(context.Background(), Config{}); err == nil {
		t.Fatalf("NewS3Store() error = nil, want error")
	}
	if _, err := NewS3Store(context.Background(), Config{Bucket: "bucket", AccessKeyID: "access", SecretAccessKey: "secret"}); err != nil {
		t.Fatalf("NewS3Store() error = %v", err)
	}
}

// TestPutHeadersReturnsContentType verifies upload headers include the required content type.
func TestPutHeadersReturnsContentType(t *testing.T) {
	headers := putHeaders("image/png")
	if headers["Content-Type"] != "image/png" {
		t.Fatalf("Content-Type = %q, want image/png", headers["Content-Type"])
	}
	if empty := putHeaders(""); empty != nil {
		t.Fatalf("putHeaders(empty) = %v, want nil", empty)
	}
}

// TestTrimETagNormalizesQuotedValues verifies ETag normalization.
func TestTrimETagNormalizesQuotedValues(t *testing.T) {
	value := `"abc"`
	if got := trimETag(&value); got != "abc" {
		t.Fatalf("trimETag() = %q, want abc", got)
	}
	if got := trimETag(nil); got != "" {
		t.Fatalf("trimETag(nil) = %q, want empty", got)
	}
}

// TestS3StorePublicURLReturnsConfiguredURL verifies public URL resolution.
func TestS3StorePublicURLReturnsConfiguredURL(t *testing.T) {
	store := S3Store{publicBaseURL: "https://cdn.test/assets"}
	if got := store.PublicURL("/folder/file.png"); got != "https://cdn.test/assets/folder/file.png" {
		t.Fatalf("PublicURL() = %q, want configured public URL", got)
	}
	if got := (S3Store{}).PublicURL("file.png"); got != "" {
		t.Fatalf("PublicURL() = %q, want empty without base URL", got)
	}
}
