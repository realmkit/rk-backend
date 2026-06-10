package redis

import (
	"context"
	"encoding/json"

	"github.com/niflaot/gamehub-go/pkg/events/domain"
	goredis "github.com/redis/go-redis/v9"
)

// DefaultChannel is the Redis pub/sub channel for events.
const DefaultChannel = "events:pubsub:v1"

// Broker publishes events to Redis pub/sub.
type Broker struct {
	client  *goredis.Client
	channel string
}

// NewBroker creates a Redis event broker.
func NewBroker(client *goredis.Client) Broker {
	return Broker{client: client, channel: DefaultChannel}
}

// WithChannel returns a broker using channel.
func (broker Broker) WithChannel(channel string) Broker {
	broker.channel = channel
	return broker
}

// Publish broadcasts one event.
func (broker Broker) Publish(ctx context.Context, event domain.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return broker.client.Publish(ctx, broker.channel, body).Err()
}
