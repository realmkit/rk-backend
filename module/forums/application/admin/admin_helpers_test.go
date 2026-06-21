package admin

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// TestSettingsPayloadRedactsExternalURL verifies settings payload shape.
func TestSettingsPayloadRedactsExternalURL(t *testing.T) {
	settings := domain.ForumSettings{ForumID: uuid.New(), ExternalURL: "https://example.test"}
	payload := settingsPayload(settings)
	if payload["external_url_configured"] != true {
		t.Fatalf("settingsPayload() = %#v, want configured URL flag", payload)
	}
	if _, exists := payload["external_url"]; exists {
		t.Fatalf("settingsPayload() leaked external_url: %#v", payload)
	}
}

// TestPermissionGrantCountSumsBuckets verifies grant bucket totals.
func TestPermissionGrantCountSumsBuckets(t *testing.T) {
	settings := domain.ForumPermissionSettings{
		Viewers:        []domain.ForumPermissionGrant{{SubjectID: uuid.New()}},
		Creators:       []domain.ForumPermissionGrant{{SubjectID: uuid.New()}, {SubjectID: uuid.New()}},
		Administrators: []domain.ForumPermissionGrant{{SubjectID: uuid.New()}},
	}
	if got := permissionGrantCount(settings); got != 4 {
		t.Fatalf("permissionGrantCount() = %d, want 4", got)
	}
}
