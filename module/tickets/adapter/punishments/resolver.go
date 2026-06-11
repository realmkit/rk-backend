// Package punishments adapts punishment services for ticket appeals.
package punishments

import (
	"context"

	"github.com/google/uuid"
	punishmentport "github.com/niflaot/gamehub-go/module/punishments/port"
	ticketport "github.com/niflaot/gamehub-go/module/tickets/port"
)

// Resolver resolves and mutates punishments for appeals.
type Resolver struct {
	service punishmentport.Service
}

// NewResolver creates a punishment resolver.
func NewResolver(service punishmentport.Service) Resolver {
	return Resolver{service: service}
}

// GetPunishment returns a safe ticket-facing punishment summary.
func (resolver Resolver) GetPunishment(ctx context.Context, id uuid.UUID) (ticketport.PunishmentSummary, error) {
	punishment, err := resolver.service.GetPunishment(ctx, id)
	if err != nil {
		return ticketport.PunishmentSummary{}, err
	}
	return ticketport.PunishmentSummary{
		ID:           punishment.ID,
		TargetUserID: punishment.TargetUserID,
		IssuerUserID: punishment.IssuerUserID,
		Status:       string(punishment.Status),
	}, nil
}

// RevokePunishment revokes one punishment after an accepted appeal.
func (resolver Resolver) RevokePunishment(ctx context.Context, id uuid.UUID, actor uuid.UUID, reason string, expected uint64) error {
	if expected == 0 {
		punishment, err := resolver.service.GetPunishment(ctx, id)
		if err != nil {
			return err
		}
		expected = punishment.Version
	}
	return resolver.service.RevokePunishment(ctx, punishmentport.RevokeCommand{
		ActorUserID:     actor,
		PunishmentID:    id,
		Reason:          reason,
		ExpectedVersion: expected,
	})
}

// Ensure Resolver implements ticket appeal ports.
var _ interface {
	ticketport.PunishmentReader
	ticketport.PunishmentExecutor
} = Resolver{}
