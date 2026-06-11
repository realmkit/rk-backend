package defaults

import (
	"time"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
)

// EventDefinitions returns event-system cron definitions.
func EventDefinitions(now time.Time) []domain.Definition {
	return []domain.Definition{
		Interval(domain.JobEventsDispatchPending, "Dispatch pending events", time.Minute, now),
	}
}
