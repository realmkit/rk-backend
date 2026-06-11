package postgres

import (
	"encoding/json"

	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
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
