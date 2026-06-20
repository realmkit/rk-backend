package css

import "testing"

// TestValidateReportsUnsafeAndInvalidCSS verifies CSS safety diagnostics.
func TestValidateReportsUnsafeAndInvalidCSS(t *testing.T) {
	issues := Validate(`@import url("https://example.test/a.css"); body { color: red;`)
	if len(issues) != 2 {
		t.Fatalf("issues = %+v, want unsafe and invalid diagnostics", issues)
	}
}

// TestValidateAcceptsBalancedLocalCSS verifies local balanced CSS passes.
func TestValidateAcceptsBalancedLocalCSS(t *testing.T) {
	issues := Validate(`body { color: var(--rk-accent); }`)
	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}
