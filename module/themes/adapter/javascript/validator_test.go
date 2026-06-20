package javascript

import "testing"

// TestValidateReportsUnsafeAndInvalidJavaScript verifies JavaScript safety diagnostics.
func TestValidateReportsUnsafeAndInvalidJavaScript(t *testing.T) {
	issues := Validate(`eval("alert(1)";`)
	if len(issues) != 2 {
		t.Fatalf("issues = %+v, want unsafe and invalid diagnostics", issues)
	}
}

// TestValidateAcceptsBalancedLocalJavaScript verifies local balanced JavaScript passes.
func TestValidateAcceptsBalancedLocalJavaScript(t *testing.T) {
	issues := Validate(`window.realmkit = { ready: true };`)
	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}
