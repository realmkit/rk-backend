package defaults

import (
	"time"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
)

// ForumDefinitions returns forum operational cron definitions.
func ForumDefinitions(now time.Time) []domain.Definition {
	return []domain.Definition{
		Interval(domain.JobForumsFlushThreadViews, "Flush forum thread views", 5*time.Minute, now),
		Interval(domain.JobForumsVerifyStats, "Verify forum stats", 24*time.Hour, now),
		Interval(domain.JobForumsVerifyLikes, "Verify forum likes", 24*time.Hour, now),
	}
}
