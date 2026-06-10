// Package testing contains reusable event test doubles.
package testing

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
)

// Recorder records published event drafts in tests.
type Recorder struct {
	mu     sync.Mutex
	drafts []domain.Draft
}

// Publish records one draft and returns a synthetic event.
func (recorder *Recorder) Publish(_ context.Context, draft domain.Draft, now time.Time) (domain.Event, error) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.drafts = append(recorder.drafts, draft)
	return domain.Event{
		ID:            uuid.New(),
		Key:           draft.Key,
		SchemaVersion: draft.SchemaVersion,
		Producer:      draft.Producer,
		AggregateType: draft.AggregateType,
		AggregateID:   draft.AggregateID,
		Status:        domain.StatusPending,
		Scopes:        draft.Scopes,
		OccurredAt:    now,
		AvailableAt:   now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Drafts returns recorded drafts.
func (recorder *Recorder) Drafts() []domain.Draft {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return append([]domain.Draft(nil), recorder.drafts...)
}
