package importing

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// fakeStore stores objects in memory.
type fakeStore struct {
	objects map[string][]byte
}

// Health reports fake storage health.
func (store *fakeStore) Health(context.Context) error {
	return nil
}

// Put stores object bytes.
func (store *fakeStore) Put(_ context.Context, object storage.Object, body io.Reader) (storage.StoredObject, error) {
	content, err := io.ReadAll(body)
	if err != nil {
		return storage.StoredObject{}, err
	}
	store.objects[object.Key] = content
	return storage.StoredObject{Key: object.Key}, nil
}

// Delete deletes one object.
func (store *fakeStore) Delete(_ context.Context, key string) error {
	delete(store.objects, key)
	return nil
}

// PresignPut is unused by importer tests.
func (store *fakeStore) PresignPut(context.Context, storage.PresignPutRequest) (storage.PresignedRequest, error) {
	return storage.PresignedRequest{}, errors.New("unused")
}

// PresignGet is unused by importer tests.
func (store *fakeStore) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "", errors.New("unused")
}

// Head is unused by importer tests.
func (store *fakeStore) Head(context.Context, string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{}, errors.New("unused")
}

// zipPackage builds a test zip package.
func zipPackage(t *testing.T, files map[string][]byte) (io.Reader, int64) {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return bytes.NewReader(buffer.Bytes()), int64(buffer.Len())
}

// issueCodes returns a set of issue codes.
func issueCodes(issues []domain.ThemeValidationIssue) map[domain.ValidationIssueCode]struct{} {
	codes := map[domain.ValidationIssueCode]struct{}{}
	for _, issue := range issues {
		codes[issue.Code] = struct{}{}
	}
	return codes
}

// assertIssue verifies an issue code exists.
func assertIssue(
	t *testing.T,
	codes map[domain.ValidationIssueCode]struct{},
	code domain.ValidationIssueCode,
) {
	t.Helper()
	if _, ok := codes[code]; !ok {
		t.Fatalf("issue %q missing from %v", code, codes)
	}
}
