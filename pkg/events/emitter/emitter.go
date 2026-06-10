package emitter

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
)

// Publisher stores one event draft.
type Publisher interface {
	// Publish stores one event draft.
	Publish(ctx context.Context, draft domain.Draft) (domain.Event, error)
}

// Publish stores one event draft when a publisher exists.
func Publish(ctx context.Context, publisher Publisher, draft domain.Draft) error {
	if publisher == nil {
		return nil
	}
	_, err := publisher.Publish(ctx, draft)
	return err
}

// UUID returns a pointer to id.
func UUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
