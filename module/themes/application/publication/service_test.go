package publication

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// TestActivatePublishesValidSignedVersion verifies publication success.
func TestActivatePublishesValidSignedVersion(t *testing.T) {
	repositories, versionID := publicationRepositories(domain.VersionStatusValid, domain.SignatureVerified)
	events := &fakeEvents{}
	service := NewService(repositories, nil, events, fixedClock())
	activation, err := service.Activate(context.Background(), ActivateCommand{
		VersionID: versionID, Environment: domain.EnvironmentPublic, SettingsDataJSON: []byte(`{"brand":"RealmKit"}`),
	})
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	version := repositories.Versions.(*fakeVersionRepository).versions[versionID]
	if version.Status != domain.VersionStatusPublished || version.PublishedAt == nil {
		t.Fatalf("version = %+v, want published", version)
	}
	if string(activation.SettingsDataJSON) != `{"brand":"RealmKit"}` {
		t.Fatalf("settings = %s, want stored settings", activation.SettingsDataJSON)
	}
	if events.calls != 1 {
		t.Fatalf("events calls = %d, want 1", events.calls)
	}
}

// TestRollbackCreatesNewActivation verifies append-only rollback behavior.
func TestRollbackCreatesNewActivation(t *testing.T) {
	repositories, versionID := publicationRepositories(domain.VersionStatusPublished, domain.SignatureVerified)
	activationID := uuid.New()
	repositories.Activations.(*fakeActivationRepository).activations[activationID] = domain.ThemeActivation{
		ID: activationID, ThemeID: uuid.New(), VersionID: versionID, Environment: domain.EnvironmentPublic, SettingsDataJSON: []byte(`{"brand":"Old"}`),
	}
	service := NewService(repositories, nil, nil, fixedClock())
	activation, err := service.Rollback(context.Background(), RollbackCommand{
		ActivationID: activationID, Environment: domain.EnvironmentPublic, Reason: "Rollback",
	})
	if err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	if activation.ID == activationID || activation.VersionID != versionID {
		t.Fatalf("activation = %+v, want new activation for previous version", activation)
	}
}

// TestActivateDenials verifies publication guard failures.
func TestActivateDenials(t *testing.T) {
	cases := []struct {
		name        string
		status      domain.VersionStatus
		signature   domain.SignatureVerificationStatus
		settings    []byte
		permissions PermissionChecker
	}{
		{name: "invalid version", status: domain.VersionStatusInvalid, signature: domain.SignatureVerified, settings: []byte(`{"brand":"RealmKit"}`)},
		{name: "revoked signature", status: domain.VersionStatusValid, signature: domain.SignatureRevoked, settings: []byte(`{"brand":"RealmKit"}`)},
		{name: "bad settings", status: domain.VersionStatusValid, signature: domain.SignatureVerified, settings: []byte(`{}`)},
		{name: "permission", status: domain.VersionStatusValid, signature: domain.SignatureVerified, settings: []byte(`{"brand":"RealmKit"}`), permissions: fakePermissions{err: port.ErrPermissionDenied}},
	}
	for _, tt := range cases {
		repositories, versionID := publicationRepositories(tt.status, tt.signature)
		service := NewService(repositories, tt.permissions, nil, fixedClock())
		_, err := service.Activate(context.Background(), ActivateCommand{
			VersionID: versionID, Environment: domain.EnvironmentPublic, SettingsDataJSON: tt.settings,
		})
		if !errors.Is(err, port.ErrInvalidState) && !errors.Is(err, port.ErrPermissionDenied) {
			t.Fatalf("%s error = %v, want invalid state or permission", tt.name, err)
		}
	}
}

// publicationRepositories returns fake publication repositories.
func publicationRepositories(
	status domain.VersionStatus,
	signatureStatus domain.SignatureVerificationStatus,
) (Repositories, uuid.UUID) {
	versionID := uuid.New()
	themeID := uuid.New()
	return Repositories{
		Versions: &fakeVersionRepository{versions: map[uuid.UUID]domain.ThemeVersion{
			versionID: {ID: versionID, ThemeID: themeID, Status: status, SettingsSchemaJSON: []byte(`{"required":["brand"]}`), Version: 1},
		}},
		Issues: &fakeIssueRepository{issues: map[uuid.UUID][]domain.ThemeValidationIssue{}},
		Signatures: &fakeSignatureRepository{signatures: map[uuid.UUID]domain.ThemePackageSignature{
			versionID: {ID: uuid.New(), VersionID: versionID, VerificationStatus: signatureStatus},
		}},
		Activations: &fakeActivationRepository{activations: map[uuid.UUID]domain.ThemeActivation{}},
	}, versionID
}

// fixedClock returns a deterministic publication clock.
func fixedClock() Clock {
	return func() time.Time { return time.Date(2026, time.June, 19, 15, 0, 0, 0, time.UTC) }
}
