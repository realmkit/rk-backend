package defaults

import (
	"time"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
)

// TicketDefinitions returns ticket operational cron definitions.
func TicketDefinitions(now time.Time) []domain.Definition {
	return []domain.Definition{
		Interval(domain.JobTicketsDetectSLABreaches, "Detect ticket SLA breaches", 15*time.Minute, now),
		Interval(domain.JobTicketsCloseStale, "Close stale tickets", 24*time.Hour, now),
		Interval(domain.JobTicketsVerifyStats, "Verify ticket stats", 24*time.Hour, now),
		Manual(domain.JobTicketsRebuildStats, "Rebuild ticket stats", now),
	}
}
