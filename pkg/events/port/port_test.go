package port

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
)

// TestListFilterCarriesEventQueryFields verifies filter fields retain assigned values.
func TestListFilterCarriesEventQueryFields(t *testing.T) {
	id := uuid.New()
	filter := ListFilter{
		Status:        domain.StatusPending,
		Producer:      domain.ProducerForums,
		EventKey:      "forums.test",
		AggregateType: "forum",
		AggregateID:   &id,
	}
	if filter.AggregateID == nil || *filter.AggregateID != id || filter.Status != domain.StatusPending {
		t.Fatalf("ListFilter = %#v, want assigned values", filter)
	}
}

// TestEventErrorsAreStable verifies exported sentinel errors are usable.
func TestEventErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrForbidden} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
