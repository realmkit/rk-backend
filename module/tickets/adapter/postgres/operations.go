package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/niflaot/gamehub-go/module/tickets/domain"
)

// VerifyStats reports ticket counter drift without mutating rows.
func (repository TicketRepository) VerifyStats(ctx context.Context) (domain.DriftReport, error) {
	var tickets []TicketModel
	if err := repository.store.DB(ctx).Find(&tickets).Error; err != nil {
		return domain.DriftReport{}, err
	}
	var mismatches []string
	for _, ticket := range tickets {
		actual, err := repository.ticketCounts(ctx, ticket.ID.ID)
		if err != nil {
			return domain.DriftReport{}, err
		}
		if actual.messages != ticket.MessageCount {
			mismatches = append(mismatches, drift(ticket.ID.ID.String(), "message_count", ticket.MessageCount, actual.messages))
		}
		if actual.staffMessages != ticket.StaffMessageCount {
			mismatches = append(mismatches, drift(ticket.ID.ID.String(), "staff_message_count", ticket.StaffMessageCount, actual.staffMessages))
		}
		if actual.evidence != ticket.EvidenceCount {
			mismatches = append(mismatches, drift(ticket.ID.ID.String(), "evidence_count", ticket.EvidenceCount, actual.evidence))
		}
	}
	return domain.DriftReport{Mismatches: mismatches}, nil
}

// RebuildStats repairs ticket counters from source rows.
func (repository TicketRepository) RebuildStats(ctx context.Context) (domain.DriftReport, error) {
	report, err := repository.VerifyStats(ctx)
	if err != nil {
		return domain.DriftReport{}, err
	}
	var tickets []TicketModel
	if err := repository.store.DB(ctx).Find(&tickets).Error; err != nil {
		return domain.DriftReport{}, err
	}
	for _, ticket := range tickets {
		counts, err := repository.ticketCounts(ctx, ticket.ID.ID)
		if err != nil {
			return domain.DriftReport{}, err
		}
		updates := map[string]any{
			"message_count":       counts.messages,
			"staff_message_count": counts.staffMessages,
			"evidence_count":      counts.evidence,
		}
		if err := repository.store.DB(ctx).Model(&TicketModel{}).
			Where("id = ?", ticket.ID.ID).Updates(updates).Error; err != nil {
			return domain.DriftReport{}, err
		}
	}
	report.Repaired = true
	return report, nil
}

// DetectSLABreaches returns open tickets whose SLA dates are overdue.
func (repository TicketRepository) DetectSLABreaches(ctx context.Context, now time.Time) ([]domain.Ticket, error) {
	statuses := []string{
		string(domain.StatusOpen),
		string(domain.StatusPendingStaff),
		string(domain.StatusPendingSubmitter),
		string(domain.StatusEscalated),
	}
	var models []TicketModel
	err := repository.store.DB(ctx).Where("status IN ? AND ((sla_first_response_due_at IS NOT NULL AND first_staff_response_at IS NULL AND sla_first_response_due_at < ?) OR (sla_resolution_due_at IS NOT NULL AND sla_resolution_due_at < ?))", statuses, now, now).
		Order("opened_at asc").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return ticketsFromModels(models), nil
}

// CloseStale closes tickets blocked on the submitter for more than fourteen days.
func (repository TicketRepository) CloseStale(ctx context.Context, now time.Time) (int64, error) {
	cutoff := now.Add(-14 * 24 * time.Hour)
	result := repository.store.DB(ctx).Model(&TicketModel{}).
		Where("status = ? AND updated_at < ?", domain.StatusPendingSubmitter, cutoff).
		Updates(map[string]any{
			"status":       string(domain.StatusClosed),
			"closed_at":    now,
			"close_reason": "closed after stale submitter inactivity",
			"resolution":   "stale",
		})
	return result.RowsAffected, result.Error
}

// ticketCounterSet contains source-of-truth aggregate counts.
type ticketCounterSet struct {
	messages      int64
	staffMessages int64
	evidence      int64
}

// ticketCounts computes counters from source rows.
func (repository TicketRepository) ticketCounts(ctx context.Context, ticketID any) (ticketCounterSet, error) {
	var counts ticketCounterSet
	db := repository.store.DB(ctx)
	if err := db.Model(&MessageModel{}).Where("ticket_id = ?", ticketID).Count(&counts.messages).Error; err != nil {
		return counts, err
	}
	if err := db.Model(&MessageModel{}).Where("ticket_id = ? AND author_role = ?", ticketID, domain.RoleStaff).Count(&counts.staffMessages).Error; err != nil {
		return counts, err
	}
	if err := db.Model(&EvidenceModel{}).Where("ticket_id = ?", ticketID).Count(&counts.evidence).Error; err != nil {
		return counts, err
	}
	return counts, nil
}

// ticketsFromModels maps ticket rows into domain tickets.
func ticketsFromModels(models []TicketModel) []domain.Ticket {
	tickets := make([]domain.Ticket, 0, len(models))
	for _, model := range models {
		tickets = append(tickets, ticketFromModel(model))
	}
	return tickets
}

// drift formats one counter mismatch.
func drift(id string, field string, stored int64, actual int64) string {
	return fmt.Sprintf("%s %s stored=%d actual=%d", id, field, stored, actual)
}
