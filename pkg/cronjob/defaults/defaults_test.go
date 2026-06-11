package defaults

import (
	"testing"
	"time"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
)

// TestOwnerDefinitionsValidate verifies code-owned owner defaults.
func TestOwnerDefinitionsValidate(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	groups := [][]domain.Definition{
		EventDefinitions(now),
		ForumDefinitions(now),
		PunishmentDefinitions(now),
		TicketDefinitions(now),
		MaintenanceDefinitions(now),
	}
	for _, group := range groups {
		for _, definition := range group {
			if err := definition.Validate(); err != nil {
				t.Fatalf("%s Validate() error = %v", definition.Key, err)
			}
		}
	}
}

// TestTicketDefinitionsContainOperationalJobs verifies ticket defaults.
func TestTicketDefinitionsContainOperationalJobs(t *testing.T) {
	required := map[string]bool{
		domain.JobTicketsDetectSLABreaches: false,
		domain.JobTicketsCloseStale:        false,
		domain.JobTicketsVerifyStats:       false,
		domain.JobTicketsRebuildStats:      false,
	}
	for _, definition := range TicketDefinitions(time.Unix(0, 0).UTC()) {
		if _, ok := required[definition.Key]; ok {
			required[definition.Key] = true
		}
	}
	for key, found := range required {
		if !found {
			t.Fatalf("TicketDefinitions() missing %s", key)
		}
	}
}
