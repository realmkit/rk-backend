package domain

import "testing"

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
