package emitter

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestPublishNoopsWithoutPublisher verifies optional publishers are safe.
func TestPublishNoopsWithoutPublisher(t *testing.T) {
	if err := Publish(context.Background(), nil, domain.Draft{}); err != nil {
		t.Fatalf("Publish() error = %v, want nil", err)
	}
}

// TestPublishDelegatesToPublisher verifies configured publishers receive drafts.
func TestPublishDelegatesToPublisher(t *testing.T) {
	publisher := &recordingPublisher{}
	draft := domain.Draft{Key: "users.user.updated"}
	if err := Publish(context.Background(), publisher, draft); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if publisher.draft.Key != draft.Key {
		t.Fatalf("draft key = %s, want %s", publisher.draft.Key, draft.Key)
	}
}

// TestUUIDReturnsNilForEmptyID verifies optional aggregate IDs omit nil UUIDs.
func TestUUIDReturnsNilForEmptyID(t *testing.T) {
	if UUID(uuid.Nil) != nil {
		t.Fatalf("UUID(nil) returned non-nil pointer")
	}
	id := uuid.New()
	if got := UUID(id); got == nil || *got != id {
		t.Fatalf("UUID() = %v, want %s", got, id)
	}
}

// recordingPublisher records one draft.
type recordingPublisher struct {
	draft domain.Draft
}

// Publish records one draft.
func (publisher *recordingPublisher) Publish(_ context.Context, draft domain.Draft) (domain.Event, error) {
	publisher.draft = draft
	return domain.Event{}, nil
}
