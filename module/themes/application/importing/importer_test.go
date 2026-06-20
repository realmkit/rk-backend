package importing

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/application/signing"
	"github.com/realmkit/rk-backend/module/themes/domain"
)

// TestServiceImportsSafePackagePersistsDraftAssets verifies successful zip import.
func TestServiceImportsSafePackagePersistsDraftAssets(t *testing.T) {
	repositories := newFakeRepositories()
	store := &fakeStore{objects: map[string][]byte{}}
	service := NewService(repositories.ports(), store, Config{}, nil)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter","version":"1.0.0"}`),
		"layout/theme.liquid": []byte(`<main>{{ content_for_layout }}</main>`),
		"assets/logo.png":     {0x89, 0x50, 0x4e, 0x47},
	})
	result, err := service.Import(context.Background(), Command{
		ThemeID:          uuid.New(),
		IdempotencyKey:   "safe-import",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !result.Imported || result.Version.Status != domain.VersionStatusDraft {
		t.Fatalf("result = %+v, want imported draft", result)
	}
	if len(repositories.files.files[result.Version.ID]) != 3 {
		t.Fatalf("files = %d, want 3", len(repositories.files.files[result.Version.ID]))
	}
	if len(repositories.assets.assets[result.Version.ID]) != 1 {
		t.Fatalf("assets = %d, want 1", len(repositories.assets.assets[result.Version.ID]))
	}
	if len(store.objects) != 1 {
		t.Fatalf("stored objects = %d, want 1", len(store.objects))
	}
}

// TestServiceReusesIdempotentImport verifies source reference retry behavior.
func TestServiceReusesIdempotentImport(t *testing.T) {
	repositories := newFakeRepositories()
	service := NewService(repositories.ports(), &fakeStore{objects: map[string][]byte{}}, Config{}, nil)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter","version":"1.0.0"}`),
	})
	themeID := uuid.New()
	first, err := service.Import(context.Background(), Command{
		ThemeID:          themeID,
		IdempotencyKey:   "retry",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("first Import() error = %v", err)
	}
	second, err := service.Import(context.Background(), Command{ThemeID: themeID, IdempotencyKey: "retry"})
	if err != nil {
		t.Fatalf("second Import() error = %v", err)
	}
	if !second.Reused || second.Version.ID != first.Version.ID {
		t.Fatalf("second = %+v, want reused first version", second)
	}
}

// TestServiceRejectsUnsafePackageIssues verifies blocking package diagnostics.
func TestServiceRejectsUnsafePackageIssues(t *testing.T) {
	repositories := newFakeRepositories()
	service := NewService(repositories.ports(), &fakeStore{objects: map[string][]byte{}}, Config{}, nil)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json":        []byte(`{"name":"Starter","version":"1.0.0"}`),
		"layout/theme.liquid":        []byte(`layout`),
		"layout/./theme.liquid":      []byte(`duplicate`),
		"templates/index.liquid":     {0xff, 0xfe},
		"../templates/escape.liquid": []byte(`escape`),
	})
	result, err := service.Import(context.Background(), Command{
		ThemeID:          uuid.New(),
		IdempotencyKey:   "unsafe",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	codes := issueCodes(result.Issues)
	assertIssue(t, codes, domain.IssueUnsafePath)
	assertIssue(t, codes, domain.IssueDuplicatePath)
	assertIssue(t, codes, domain.IssueInvalidUTF8)
	if result.Version.Status != domain.VersionStatusInvalid {
		t.Fatalf("Status = %q, want invalid", result.Version.Status)
	}
	if len(repositories.files.files[result.Version.ID]) != 0 {
		t.Fatalf("files persisted = %d, want 0", len(repositories.files.files[result.Version.ID]))
	}
}

// TestExtractPackageEnforcesLimits verifies size, count, and compression diagnostics.
func TestExtractPackageEnforcesLimits(t *testing.T) {
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter"}`),
		"layout/theme.liquid": bytes.Repeat([]byte("a"), 128),
	})
	_, issues, err := extractPackage(context.Background(), reader, size, Config{
		MaxPackageBytes:     DefaultMaxPackageBytes,
		MaxExtractedBytes:   DefaultMaxExtractedBytes,
		MaxFileCount:        1,
		MaxTextFileBytes:    8,
		MaxCompressionRatio: 0.1,
	}.Defaults())
	if err != nil {
		t.Fatalf("extractPackage() error = %v", err)
	}
	codes := issueCodes(issues)
	assertIssue(t, codes, domain.IssueFileCountTooLarge)
	assertIssue(t, codes, domain.IssueTextFileTooLarge)
	assertIssue(t, codes, domain.IssueCompressionRatioTooLarge)
	_, packageIssues, err := extractPackage(context.Background(), bytes.NewReader([]byte("too large")), 99, Config{MaxPackageBytes: 4}.Defaults())
	if err != nil {
		t.Fatalf("extractPackage(large) error = %v", err)
	}
	assertIssue(t, issueCodes(packageIssues), domain.IssuePackageTooLarge)
}

// TestExtractPackageHonorsCancellation verifies extraction exits on cancellation.
func TestExtractPackageHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter"}`),
	})

	_, _, err := extractPackage(ctx, reader, size, Config{}.Defaults())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("extractPackage() error = %v, want canceled", err)
	}
}

// TestServicePersistsSignatureResult verifies verifier output persistence.
func TestServicePersistsSignatureResult(t *testing.T) {
	repositories := newFakeRepositories()
	verifier := &fakeVerifier{result: signing.Result{
		Signature: domain.ThemePackageSignature{
			ID:                 uuid.New(),
			Algorithm:          domain.SignatureAlgorithmEd25519,
			VerificationStatus: domain.SignatureVerified,
		},
	}}
	service := NewService(repositories.ports(), &fakeStore{objects: map[string][]byte{}}, Config{}, verifier)
	reader, size := zipPackage(t, map[string][]byte{
		"realmkit-theme.json": []byte(`{"name":"Starter"}`),
	})
	result, err := service.Import(context.Background(), Command{
		ThemeID:          uuid.New(),
		IdempotencyKey:   "signed",
		PackageSizeBytes: size,
		Package:          reader,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	signature := repositories.signatures.signatures[result.Version.ID]
	if signature.VerificationStatus != domain.SignatureVerified {
		t.Fatalf("signature = %+v, want verified", signature)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
}
