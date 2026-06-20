package http

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
)

// TestHubAddRemovePublishAndMatch covers local hub client bookkeeping.
func TestHubAddRemovePublishAndMatch(t *testing.T) {
	hub := NewHub()
	userID := uuid.New()
	eventScope := domain.Scope{Type: domain.ScopeUser, ID: userID.String()}
	client := &client{
		id: uuid.New(),
		scopes: map[string]domain.Scope{
			scopeKey(domain.Scope{Type: domain.ScopeGlobal}): {Type: domain.ScopeGlobal},
		},
	}

	hub.add(client)
	if got := len(hub.clients); got != 1 {
		t.Fatalf("client count = %d, want 1", got)
	}
	if client.matches([]domain.Scope{eventScope}) {
		t.Fatalf("expected client not to match unsubscribed user scope")
	}
	if !client.matches([]domain.Scope{{Type: domain.ScopeGlobal}, eventScope}) {
		t.Fatalf("expected client to match global scope")
	}

	event := domain.Event{
		ID:         uuid.New(),
		Key:        domain.EventForumsThreadCreated,
		Producer:   domain.ProducerForums,
		Scopes:     []domain.Scope{eventScope},
		OccurredAt: time.Now().UTC(),
	}
	if err := hub.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	hub.Broadcast(context.Background(), event)

	hub.remove(client.id)
	if got := len(hub.clients); got != 0 {
		t.Fatalf("client count after remove = %d, want 0", got)
	}
}
