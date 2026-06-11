package defaults

import (
	"time"

	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
)

// PunishmentDefinitions returns punishment operational cron definitions.
func PunishmentDefinitions(now time.Time) []domain.Definition {
	return []domain.Definition{
		Interval(domain.JobPunishmentsExpireActive, "Expire active punishments", 5*time.Minute, now),
		Interval(domain.JobPunishmentsVerifyRestrictions, "Verify punishment restrictions", 24*time.Hour, now),
		Manual(domain.JobPunishmentsRebuildRestrictions, "Rebuild punishment restrictions", now),
	}
}
