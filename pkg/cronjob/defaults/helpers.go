// Package defaults provides code-owned initial cron job schedules.
package defaults

import (
	"time"

	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
)

// Interval creates an enabled interval definition.
func Interval(key string, name string, duration time.Duration, now time.Time) domain.Definition {
	next := now.Add(duration)
	return domain.Definition{
		Key:                key,
		Name:               name,
		ScheduleKind:       domain.ScheduleInterval,
		ScheduleExpression: duration.String(),
		Enabled:            true,
		ConcurrencyPolicy:  domain.ConcurrencyForbid,
		NextRunAt:          &next,
		Version:            1,
	}
}

// Manual creates a disabled manual definition.
func Manual(key string, name string, now time.Time) domain.Definition {
	return domain.Definition{
		Key:                key,
		Name:               name,
		ScheduleKind:       domain.ScheduleManual,
		ScheduleExpression: "",
		Enabled:            false,
		ConcurrencyPolicy:  domain.ConcurrencyForbid,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
