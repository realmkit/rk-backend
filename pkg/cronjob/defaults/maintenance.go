package defaults

import (
	"time"

	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
)

// MaintenanceDefinitions returns platform maintenance cron definitions.
func MaintenanceDefinitions(now time.Time) []domain.Definition {
	return []domain.Definition{
		Interval(domain.JobAssetsExpireUploadIntents, "Expire upload intents", 30*time.Minute, now),
		Interval(domain.JobUsersCleanupIdentityClaims, "Cleanup identity claims", 24*time.Hour, now),
	}
}
