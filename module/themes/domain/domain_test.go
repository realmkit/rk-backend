package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestThemePermissionsReturnsStableCatalog verifies all theme permissions are listed.
func TestThemePermissionsReturnsStableCatalog(t *testing.T) {
	permissions := ThemePermissions()
	want := []Permission{
		PermissionThemesView,
		PermissionThemesImport,
		PermissionThemesEdit,
		PermissionThemesValidate,
		PermissionThemesPublish,
		PermissionThemesRollback,
		PermissionThemesDelete,
		PermissionThemesPreview,
		PermissionThemesActivate,
	}
	assertEqualList(t, "ThemePermissions", permissions, want)
}

// TestThemeEventKeysReturnsStableCatalog verifies all emitted theme events are listed.
func TestThemeEventKeysReturnsStableCatalog(t *testing.T) {
	keys := ThemeEventKeys()
	want := []EventKey{
		EventThemeCreated,
		EventThemeUpdated,
		EventVersionImported,
		EventVersionValidated,
		EventVersionFileSaved,
		EventVersionArchived,
		EventActivationChanged,
		EventActivationRolledBack,
		EventSigningKeyCreated,
		EventSigningKeyRetired,
		EventSigningKeyRevoked,
		EventCacheInvalidated,
	}
	assertEqualList(t, "ThemeEventKeys", keys, want)
}

// TestRouteKindsReturnsStableCatalog verifies first-version route contracts are listed.
func TestRouteKindsReturnsStableCatalog(t *testing.T) {
	routes := RouteKinds()
	want := []RouteKind{
		RouteHome,
		RouteForumsIndex,
		RouteForumsCategory,
		RouteForumsShow,
		RouteThreadsShow,
		RouteThreadsNew,
		RouteTicketsIndex,
		RouteTicketsNew,
		RouteTicketsShow,
		RoutePunishmentsIndex,
		RoutePunishmentsShow,
		RouteUsersShow,
		RouteSearch,
		RouteStaticPage,
		RouteNotFound,
		RouteError,
		RouteMaintenance,
		RouteLogin,
		RouteRegister,
		RouteForgotPassword,
		RouteResetPassword,
		RouteVerifyEmail,
		RouteAccountRecovery,
	}
	assertEqualList(t, "RouteKinds", routes, want)
}

// TestValidationIssueCodesReturnsStableCatalog verifies structured diagnostics are listed.
func TestValidationIssueCodesReturnsStableCatalog(t *testing.T) {
	codes := ValidationIssueCodes()
	assertUnique(t, "ValidationIssueCodes", codes)
	if len(codes) < 20 {
		t.Fatalf("ValidationIssueCodes() returned %d codes, want broad first-version coverage", len(codes))
	}
}

// TestCalculateVersionIntegritySHA256IsDeterministic verifies file order does not affect digests.
func TestCalculateVersionIntegritySHA256IsDeterministic(t *testing.T) {
	left := CalculateVersionIntegritySHA256([]IntegrityFile{
		{Path: "assets\\app.css", ContentSHA256: "A"},
		{Path: "layout/theme.liquid", ContentSHA256: "B"},
	})
	right := CalculateVersionIntegritySHA256([]IntegrityFile{
		{Path: "/layout/theme.liquid", ContentSHA256: "b"},
		{Path: "assets/app.css", ContentSHA256: "a"},
	})
	if left != right {
		t.Fatalf("CalculateVersionIntegritySHA256() mismatch = %q vs %q", left, right)
	}
	if left == "" {
		t.Fatalf("CalculateVersionIntegritySHA256() returned empty digest")
	}
}

// TestThemeVersionEnsureEditableRejectsPublished verifies immutable published versions.
func TestThemeVersionEnsureEditableRejectsPublished(t *testing.T) {
	if err := (ThemeVersion{Status: VersionStatusDraft}).EnsureEditable(); err != nil {
		t.Fatalf("EnsureEditable(draft) error = %v", err)
	}
	publishedAt := time.Now()
	err := (ThemeVersion{Status: VersionStatusValid, PublishedAt: &publishedAt}).EnsureEditable()
	if !errors.Is(err, ErrPublishedVersionImmutable) {
		t.Fatalf("EnsureEditable(published_at) error = %v, want %v", err, ErrPublishedVersionImmutable)
	}
	err = (ThemeVersion{Status: VersionStatusPublished}).EnsureEditable()
	if !errors.Is(err, ErrPublishedVersionImmutable) {
		t.Fatalf("EnsureEditable(published status) error = %v, want %v", err, ErrPublishedVersionImmutable)
	}
}

// TestThemeVersionEnsurePublishableRequiresValidSignatureAndNoErrors verifies activation invariants.
func TestThemeVersionEnsurePublishableRequiresValidSignatureAndNoErrors(t *testing.T) {
	version := ThemeVersion{Status: VersionStatusValid}
	signature := ThemePackageSignature{VerificationStatus: SignatureVerified}
	if err := version.EnsurePublishable(signature, nil); err != nil {
		t.Fatalf("EnsurePublishable(valid) error = %v", err)
	}
	issue := ThemeValidationIssue{Severity: SeverityError}
	if err := version.EnsurePublishable(signature, []ThemeValidationIssue{issue}); !errors.Is(err, ErrVersionNotPublishable) {
		t.Fatalf("EnsurePublishable(error issue) error = %v, want %v", err, ErrVersionNotPublishable)
	}
	unsigned := ThemePackageSignature{VerificationStatus: SignatureMissing}
	if err := version.EnsurePublishable(unsigned, nil); !errors.Is(err, ErrVersionNotPublishable) {
		t.Fatalf("EnsurePublishable(unsigned) error = %v, want %v", err, ErrVersionNotPublishable)
	}
	draft := ThemeVersion{Status: VersionStatusDraft}
	if err := draft.EnsurePublishable(signature, nil); !errors.Is(err, ErrVersionNotPublishable) {
		t.Fatalf("EnsurePublishable(draft) error = %v, want %v", err, ErrVersionNotPublishable)
	}
}

// TestThemeVersionMarkPublishedSetsAuditFields verifies publish state mutation.
func TestThemeVersionMarkPublishedSetsAuditFields(t *testing.T) {
	actor := uuid.New()
	now := time.Date(2026, time.June, 20, 12, 0, 0, 0, time.FixedZone("test", -5*60*60))
	version, err := (ThemeVersion{Status: VersionStatusValid}).MarkPublished(now, &actor)
	if err != nil {
		t.Fatalf("MarkPublished() error = %v", err)
	}
	if version.Status != VersionStatusPublished || version.PublishedAt == nil || !version.PublishedAt.Equal(now.UTC()) {
		t.Fatalf("MarkPublished() version = %#v, want published at UTC time", version)
	}
	if version.PublishedBy == nil || *version.PublishedBy != actor || version.UpdatedBy == nil || *version.UpdatedBy != actor {
		t.Fatalf("MarkPublished() actor fields = %#v", version)
	}
}

// TestThemeActivationValidateRejectsIncompleteState verifies activation invariants.
func TestThemeActivationValidateRejectsIncompleteState(t *testing.T) {
	valid := ThemeActivation{
		ThemeID:     uuid.New(),
		VersionID:   uuid.New(),
		Environment: EnvironmentPublic,
		ActivatedAt: time.Now(),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate(valid activation) error = %v", err)
	}
	invalid := valid
	invalid.Environment = ""
	if err := invalid.Validate(); !errors.Is(err, ErrInvalidActivation) {
		t.Fatalf("Validate(invalid activation) error = %v, want %v", err, ErrInvalidActivation)
	}
}

// TestThemeSigningKeyTransitionsPreserveLifecycleInvariants verifies key state changes.
func TestThemeSigningKeyTransitionsPreserveLifecycleInvariants(t *testing.T) {
	now := time.Now().UTC()
	key := ThemeSigningKey{Status: SigningKeyTrusted}
	if err := key.EnsureUsableAt(now); err != nil {
		t.Fatalf("EnsureUsableAt(trusted) error = %v", err)
	}
	retired, err := key.Retire(now)
	if err != nil {
		t.Fatalf("Retire() error = %v", err)
	}
	if retired.Status != SigningKeyRetired || retired.RetiredAt == nil {
		t.Fatalf("Retire() key = %#v, want retired timestamp", retired)
	}
	revoked := retired.Revoke(now.Add(time.Minute))
	if err := revoked.EnsureUsableAt(now); !errors.Is(err, ErrSigningKeyInactive) {
		t.Fatalf("EnsureUsableAt(revoked) error = %v, want %v", err, ErrSigningKeyInactive)
	}
	if _, err := revoked.Retire(now); !errors.Is(err, ErrSigningKeyInactive) {
		t.Fatalf("Retire(revoked) error = %v, want %v", err, ErrSigningKeyInactive)
	}
}

// TestThemePreviewTokenValidateAtChecksRevocationAndExpiry verifies preview token state.
func TestThemePreviewTokenValidateAtChecksRevocationAndExpiry(t *testing.T) {
	now := time.Now().UTC()
	token := ThemePreviewToken{VersionID: uuid.New(), TokenHash: "hash", ExpiresAt: now.Add(time.Minute)}
	if err := token.ValidateAt(now); err != nil {
		t.Fatalf("ValidateAt(valid) error = %v", err)
	}
	expired := token
	expired.ExpiresAt = now
	if err := expired.ValidateAt(now); !errors.Is(err, ErrPreviewTokenExpired) {
		t.Fatalf("ValidateAt(expired) error = %v, want %v", err, ErrPreviewTokenExpired)
	}
	revoked := token.Revoke(now)
	if err := revoked.ValidateAt(now); !errors.Is(err, ErrPreviewTokenRevoked) {
		t.Fatalf("ValidateAt(revoked) error = %v, want %v", err, ErrPreviewTokenRevoked)
	}
}

// assertEqualList compares two same-typed slices.
func assertEqualList[T comparable](t *testing.T, name string, got []T, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s() length = %d, want %d", name, len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("%s()[%d] = %v, want %v", name, index, got[index], want[index])
		}
	}
	assertUnique(t, name, got)
}

// assertUnique verifies each value appears once.
func assertUnique[T comparable](t *testing.T, name string, values []T) {
	t.Helper()
	seen := make(map[T]bool, len(values))
	for _, value := range values {
		if seen[value] {
			t.Fatalf("%s() contains duplicate %v", name, value)
		}
		seen[value] = true
	}
}
