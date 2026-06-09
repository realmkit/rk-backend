package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestUserValidateAndAuthenticate verifies user validation and auth state.
func TestUserValidateAndAuthenticate(t *testing.T) {
	user := User{ID: uuid.New(), Status: StatusActive, FirstSeenAt: time.Now().UTC()}
	if err := user.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !user.CanAuthenticate() {
		t.Fatalf("CanAuthenticate() = false, want true")
	}
	if (User{Status: StatusDisabled}).CanAuthenticate() {
		t.Fatalf("CanAuthenticate() = true, want false for disabled")
	}
}

// TestUserValidateRejectsInvalidStatus verifies invalid user status.
func TestUserValidateRejectsInvalidStatus(t *testing.T) {
	err := User{Status: "missing"}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 2 {
		t.Fatalf("Violations = %d, want at least 2", len(validation.Violations))
	}
}

// TestIdentityLinkValidateRejectsMissingKeys verifies identity link validation.
func TestIdentityLinkValidateRejectsMissingKeys(t *testing.T) {
	err := IdentityLink{}.Validate()
	var validation ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) < 5 {
		t.Fatalf("Violations = %d, want at least 5", len(validation.Violations))
	}
}
