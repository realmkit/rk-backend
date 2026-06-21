package admin

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// benchmarkSettingsPayload stores the admin settings payload benchmark result.
var benchmarkSettingsPayload map[string]any

// benchmarkPermissionGrantCount stores the permission count benchmark result.
var benchmarkPermissionGrantCount int

// BenchmarkSettingsPayload measures safe forum settings event payload assembly.
func BenchmarkSettingsPayload(b *testing.B) {
	settings := domain.ForumSettings{
		ForumID:                       uuid.New(),
		Kind:                          domain.ForumKindDiscussion,
		ThreadVisibilityMode:          domain.ThreadVisibilityAllThreads,
		MaxStickyThreads:              5,
		DefaultThreadStatus:           domain.ThreadStatusOpen,
		AuthorPostEditWindowSeconds:   600,
		AuthorPostDeleteWindowSeconds: 300,
		ExternalURL:                   "https://example.test/private",
		Version:                       3,
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkSettingsPayload = settingsPayload(settings)
	}
}

// BenchmarkPermissionGrantCount measures aggregate grant counting across every permission bucket.
func BenchmarkPermissionGrantCount(b *testing.B) {
	settings := domain.ForumPermissionSettings{
		Viewers:          benchmarkGrants(8),
		Creators:         benchmarkGrants(4),
		Replyers:         benchmarkGrants(4),
		Likers:           benchmarkGrants(4),
		ThreadPinners:    benchmarkGrants(2),
		ThreadManagers:   benchmarkGrants(2),
		PostManagers:     benchmarkGrants(2),
		LimitBypassers:   benchmarkGrants(2),
		AllThreadViewers: benchmarkGrants(2),
		Administrators:   benchmarkGrants(2),
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkPermissionGrantCount = permissionGrantCount(settings)
	}
}

// benchmarkGrants returns forum permission grants for benchmark fixtures.
func benchmarkGrants(count int) []domain.ForumPermissionGrant {
	grants := make([]domain.ForumPermissionGrant, count)
	for index := range grants {
		grants[index] = domain.ForumPermissionGrant{SubjectID: uuid.New()}
	}
	return grants
}
