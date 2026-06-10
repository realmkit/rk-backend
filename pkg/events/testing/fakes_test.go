package testing

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestRecorderStoresDrafts verifies the reusable publisher fake.
func TestRecorderStoresDrafts(t *testing.T) {
	recorder := &Recorder{}
	draft := domain.Draft{
		Key:           domain.EventForumsThreadCreated,
		SchemaVersion: 1,
		Producer:      domain.ProducerForums,
		AggregateType: "forum_thread",
		Scopes:        []domain.Scope{{Type: domain.ScopeStaff}},
	}

	event, err := recorder.Publish(context.Background(), draft, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if event.Key != draft.Key || len(recorder.Drafts()) != 1 {
		t.Fatalf("event=%+v drafts=%d, want one recorded draft", event, len(recorder.Drafts()))
	}
}
