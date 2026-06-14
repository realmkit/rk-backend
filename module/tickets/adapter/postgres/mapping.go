package postgres

import (
	"encoding/json"

	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/search"
)

// definitionModel maps a domain definition into persistence state.
func definitionModel(definition domain.Definition) DefinitionModel {
	return DefinitionModel{
		ID:                      orm.ID{ID: definition.ID},
		Key:                     string(definition.Key),
		Name:                    definition.Name,
		Description:             definition.Description,
		Kind:                    string(definition.Kind),
		Status:                  string(definition.Status),
		DefaultTeamGroupID:      definition.DefaultTeamGroupID,
		DefaultAssigneeUserID:   definition.DefaultAssigneeUserID,
		SubmitterCanClose:       definition.SubmitterCanClose,
		SubmitterCanReopen:      definition.SubmitterCanReopen,
		AllowAnonymousSubmitter: definition.AllowAnonymousSubmitter,
		RequiresTargetUser:      definition.RequiresTargetUser,
		RequiresPunishment:      definition.RequiresPunishment,
		RequiresEvidence:        definition.RequiresEvidence,
		MaxOpenPerSubmitter:     definition.MaxOpenPerSubmitter,
		ReopenWindowSeconds:     definition.ReopenWindowSeconds,
		SLAFirstResponseSeconds: definition.SLAFirstResponseSeconds,
		SLAResolutionSeconds:    definition.SLAResolutionSeconds,
		MetadataSchemaKey:       definition.MetadataSchemaKey,
		DisplayOrder:            definition.DisplayOrder,
		Version:                 definition.Version,
	}
}

// definitionFromModel maps persistence rows into domain definitions.
func definitionFromModel(model DefinitionModel) domain.Definition {
	return domain.Definition{
		ID:                      model.ID.ID,
		Key:                     domain.Key(model.Key),
		Name:                    model.Name,
		Description:             model.Description,
		Kind:                    domain.Kind(model.Kind),
		Status:                  domain.DefinitionStatus(model.Status),
		DefaultTeamGroupID:      model.DefaultTeamGroupID,
		DefaultAssigneeUserID:   model.DefaultAssigneeUserID,
		SubmitterCanClose:       model.SubmitterCanClose,
		SubmitterCanReopen:      model.SubmitterCanReopen,
		AllowAnonymousSubmitter: model.AllowAnonymousSubmitter,
		RequiresTargetUser:      model.RequiresTargetUser,
		RequiresPunishment:      model.RequiresPunishment,
		RequiresEvidence:        model.RequiresEvidence,
		MaxOpenPerSubmitter:     model.MaxOpenPerSubmitter,
		ReopenWindowSeconds:     model.ReopenWindowSeconds,
		SLAFirstResponseSeconds: model.SLAFirstResponseSeconds,
		SLAResolutionSeconds:    model.SLAResolutionSeconds,
		MetadataSchemaKey:       model.MetadataSchemaKey,
		DisplayOrder:            model.DisplayOrder,
		Version:                 model.Version,
		CreatedAt:               model.CreatedAt,
		UpdatedAt:               model.UpdatedAt,
	}
}

// ticketModel maps a domain ticket into persistence state.
func ticketModel(ticket domain.Ticket) TicketModel {
	return TicketModel{
		ID:                      orm.ID{ID: ticket.ID},
		DefinitionID:            ticket.DefinitionID,
		Key:                     ticket.Key,
		Title:                   ticket.Title,
		Kind:                    string(ticket.Kind),
		Status:                  string(ticket.Status),
		Priority:                string(ticket.Priority),
		SubmitterUserID:         ticket.SubmitterUserID,
		TargetUserID:            ticket.TargetUserID,
		PunishmentID:            ticket.PunishmentID,
		CurrentTeamGroupID:      ticket.CurrentTeamGroupID,
		AssigneeUserID:          ticket.AssigneeUserID,
		OpenedAt:                ticket.OpenedAt,
		FirstStaffResponseAt:    ticket.FirstStaffResponseAt,
		LastMessageAt:           ticket.LastMessageAt,
		LastMessageAuthorUserID: ticket.LastMessageAuthorUserID,
		ClosedAt:                ticket.ClosedAt,
		ClosedByUserID:          ticket.ClosedByUserID,
		CloseReason:             ticket.CloseReason,
		Resolution:              ticket.Resolution,
		EscalationLevel:         ticket.EscalationLevel,
		SLAFirstResponseDueAt:   ticket.SLAFirstResponseDueAt,
		SLAResolutionDueAt:      ticket.SLAResolutionDueAt,
		MessageCount:            ticket.MessageCount,
		StaffMessageCount:       ticket.StaffMessageCount,
		EvidenceCount:           ticket.EvidenceCount,
		Version:                 ticket.Version,
	}
}

// ticketFromModel maps persistence rows into domain tickets.
func ticketFromModel(model TicketModel) domain.Ticket {
	return domain.Ticket{
		ID:                      model.ID.ID,
		DefinitionID:            model.DefinitionID,
		Key:                     model.Key,
		Title:                   model.Title,
		Kind:                    domain.Kind(model.Kind),
		Status:                  domain.TicketStatus(model.Status),
		Priority:                domain.Priority(model.Priority),
		SubmitterUserID:         model.SubmitterUserID,
		TargetUserID:            model.TargetUserID,
		PunishmentID:            model.PunishmentID,
		CurrentTeamGroupID:      model.CurrentTeamGroupID,
		AssigneeUserID:          model.AssigneeUserID,
		OpenedAt:                model.OpenedAt,
		FirstStaffResponseAt:    model.FirstStaffResponseAt,
		LastMessageAt:           model.LastMessageAt,
		LastMessageAuthorUserID: model.LastMessageAuthorUserID,
		ClosedAt:                model.ClosedAt,
		ClosedByUserID:          model.ClosedByUserID,
		CloseReason:             model.CloseReason,
		Resolution:              model.Resolution,
		EscalationLevel:         model.EscalationLevel,
		SLAFirstResponseDueAt:   model.SLAFirstResponseDueAt,
		SLAResolutionDueAt:      model.SLAResolutionDueAt,
		MessageCount:            model.MessageCount,
		StaffMessageCount:       model.StaffMessageCount,
		EvidenceCount:           model.EvidenceCount,
		Version:                 model.Version,
		CreatedAt:               model.CreatedAt,
		UpdatedAt:               model.UpdatedAt,
	}
}

// ticketUpdates returns mutable update columns.
func ticketUpdates(model TicketModel, expectedVersion uint64) map[string]any {
	return map[string]any{
		"title":                       model.Title,
		"status":                      model.Status,
		"priority":                    model.Priority,
		"target_user_id":              model.TargetUserID,
		"punishment_id":               model.PunishmentID,
		"current_team_group_id":       model.CurrentTeamGroupID,
		"assignee_user_id":            model.AssigneeUserID,
		"first_staff_response_at":     model.FirstStaffResponseAt,
		"last_message_at":             model.LastMessageAt,
		"last_message_author_user_id": model.LastMessageAuthorUserID,
		"closed_at":                   model.ClosedAt,
		"closed_by_user_id":           model.ClosedByUserID,
		"close_reason":                model.CloseReason,
		"resolution":                  model.Resolution,
		"escalation_level":            model.EscalationLevel,
		"sla_first_response_due_at":   model.SLAFirstResponseDueAt,
		"sla_resolution_due_at":       model.SLAResolutionDueAt,
		"message_count":               model.MessageCount,
		"staff_message_count":         model.StaffMessageCount,
		"evidence_count":              model.EvidenceCount,
		"version":                     expectedVersion + 1,
	}
}

// ticketSearchCondition returns searchable ticket queue fields.
func ticketSearchCondition() string {
	return "LOWER(title) LIKE ? OR LOWER(CAST(id AS text)) LIKE ? OR LOWER(CAST(target_user_id AS text)) LIKE ? OR LOWER(CAST(punishment_id AS text)) LIKE ?"
}

// ticketPostgresSearchCondition returns indexed PostgreSQL text search.
func ticketPostgresSearchCondition() string {
	return "to_tsvector('simple', coalesce(title, '') || ' ' || id::text || ' ' || coalesce(target_user_id::text, '') || ' ' || coalesce(punishment_id::text, '')) @@ plainto_tsquery('simple', ?) OR id::text ILIKE ?"
}

// ticketFilterHash binds cursors to active ticket filters.
func ticketFilterHash(filter port.TicketFilter, sort search.Sort) string {
	return search.HashFilter(
		filter.SubmitterUserID,
		filter.TargetUserID,
		filter.PunishmentID,
		filter.CurrentTeamGroupID,
		filter.AssigneeUserID,
		filter.Status,
		filter.Kind,
		filter.Query.String(),
		sort,
	)
}

// definitionFilterHash binds cursors to active ticket definition filters.
func definitionFilterHash(filter port.DefinitionFilter, sort search.Sort) string {
	return search.HashFilter(filter.Kind, filter.Status, filter.Query.String(), sort)
}

// messageModel maps a domain message into persistence state.
func messageModel(message domain.Message) MessageModel {
	return MessageModel{
		ID:                  orm.ID{ID: message.ID},
		TicketID:            message.TicketID,
		AuthorUserID:        message.AuthorUserID,
		AuthorRole:          string(message.AuthorRole),
		Visibility:          string(message.Visibility),
		Sequence:            message.Sequence,
		ContentFormat:       message.ContentFormat,
		ContentDocumentJSON: string(message.ContentDocumentJSON),
		ContentText:         message.ContentText,
		ContentChecksum:     message.ContentChecksum,
		EditCount:           message.EditCount,
		Version:             message.Version,
	}
}

// messageFromModel maps persistence rows into domain messages.
func messageFromModel(model MessageModel) domain.Message {
	return domain.Message{
		ID:                  model.ID.ID,
		TicketID:            model.TicketID,
		AuthorUserID:        model.AuthorUserID,
		AuthorRole:          domain.AuthorRole(model.AuthorRole),
		Visibility:          domain.MessageVisibility(model.Visibility),
		Sequence:            model.Sequence,
		ContentFormat:       model.ContentFormat,
		ContentDocumentJSON: json.RawMessage(model.ContentDocumentJSON),
		ContentText:         model.ContentText,
		ContentChecksum:     model.ContentChecksum,
		EditCount:           model.EditCount,
		Version:             model.Version,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}
