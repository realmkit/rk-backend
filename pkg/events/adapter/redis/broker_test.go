package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	goredis "github.com/redis/go-redis/v9"
)

// TestBrokerPublishSendsPubSubMessage verifies Redis fanout.
func TestBrokerPublishSendsPubSubMessage(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()
	sub := client.Subscribe(context.Background(), DefaultChannel)
	defer sub.Close()
	if _, err := sub.Receive(context.Background()); err != nil {
		t.Fatalf("Receive subscription error = %v", err)
	}

	event := domain.Event{
		ID:            uuid.New(),
		Key:           domain.EventForumsThreadCreated,
		SchemaVersion: 1,
		Producer:      domain.ProducerForums,
		AggregateType: "forum_thread",
		Status:        domain.StatusPending,
		Scopes:        []domain.Scope{{Type: domain.ScopeStaff}},
		OccurredAt:    time.Unix(0, 0).UTC(),
		AvailableAt:   time.Unix(0, 0).UTC(),
	}
	if err := NewBroker(client).Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	message, err := sub.ReceiveMessage(context.Background())
	if err != nil {
		t.Fatalf("ReceiveMessage() error = %v", err)
	}
	if message.Payload == "" {
		t.Fatalf("message payload is empty")
	}
}
