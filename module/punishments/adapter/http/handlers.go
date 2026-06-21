package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/search"
)

// definitionRequest defines package data.
type definitionRequest struct {
	Key                    domain.Key              `json:"key"`                      // Key stores the key value.
	Name                   string                  `json:"name"`                     // Name stores the name value.
	Description            string                  `json:"description"`              // Description stores the description value.
	Color                  domain.Color            `json:"color"`                    // Color stores the color value.
	Severity               int                     `json:"severity"`                 // Severity stores the severity value.
	Status                 domain.DefinitionStatus `json:"status"`                   // Status stores the status value.
	DefaultDurationSeconds *int64                  `json:"default_duration_seconds"` // DefaultDurationSeconds stores the default duration seconds value.
	MinDurationSeconds     *int64                  `json:"min_duration_seconds"`     // MinDurationSeconds stores the min duration seconds value.
	MaxDurationSeconds     *int64                  `json:"max_duration_seconds"`     // MaxDurationSeconds stores the max duration seconds value.
	AllowPermanent         bool                    `json:"allow_permanent"`          // AllowPermanent stores the allow permanent value.
	RequiresReason         bool                    `json:"requires_reason"`          // RequiresReason stores the requires reason value.
	RequiresTargetIP       bool                    `json:"requires_target_ip"`       // RequiresTargetIP stores the requires target i p value.
	DisplayOrder           int                     `json:"display_order"`            // DisplayOrder stores the display order value.
	Actions                []domain.ActionTemplate `json:"actions"`                  // Actions stores the actions value.
}

// createDefinition supports package behavior.
func (handler handler) createDefinition(ctx *fiber.Ctx) error {
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsManageDefinitions, uuid.Nil); err != nil {
		return err
	}
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	created, err := handler.services.Punishments.CreateDefinition(ctx.UserContext(), definitionFromRequest(request, uuid.Nil))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, created.Version)
	return writeJSON(ctx, fiber.StatusCreated, created)
}

// listDefinitions supports package behavior.
func (handler handler) listDefinitions(ctx *fiber.Ctx) error {
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsManageDefinitions, uuid.Nil); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := definitionFilter(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Punishments.ListDefinitions(
		ctx.UserContext(),
		filter,
		page,
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// getDefinition supports package behavior.
func (handler handler) getDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsManageDefinitions, id); err != nil {
		return err
	}
	definition, err := handler.services.Punishments.GetDefinition(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, definition.Version)
	return writeJSON(ctx, fiber.StatusOK, definition)
}

// updateDefinition supports package behavior.
func (handler handler) updateDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsManageDefinitions, id); err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request definitionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	updated, err := handler.services.Punishments.UpdateDefinition(ctx.UserContext(), definitionFromRequest(request, id), version)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// deleteDefinition supports package behavior.
func (handler handler) deleteDefinition(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsManageDefinitions, id); err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Punishments.DeleteDefinition(ctx.UserContext(), id, version); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// reorderActions supports package behavior.
func (handler handler) reorderActions(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "definition_id")
	if err != nil {
		return err
	}
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsManageDefinitions, id); err != nil {
		return err
	}
	var request struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	if err := handler.services.Punishments.ReorderDefinitionActions(ctx.UserContext(), id, request.IDs); err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// definitionFromRequest maps HTTP definition input to domain state.
func definitionFromRequest(request definitionRequest, id uuid.UUID) domain.Definition {
	return domain.Definition{
		ID:                     id,
		Key:                    request.Key,
		Name:                   request.Name,
		Description:            request.Description,
		Color:                  request.Color,
		Severity:               request.Severity,
		Status:                 request.Status,
		DefaultDurationSeconds: request.DefaultDurationSeconds,
		MinDurationSeconds:     request.MinDurationSeconds,
		MaxDurationSeconds:     request.MaxDurationSeconds,
		AllowPermanent:         request.AllowPermanent,
		RequiresReason:         request.RequiresReason,
		RequiresTargetIP:       request.RequiresTargetIP,
		DisplayOrder:           request.DisplayOrder,
		Actions:                request.Actions,
	}
}

// definitionFilter maps list query parameters to a definition filter.
func definitionFilter(ctx *fiber.Ctx) (port.DefinitionFilter, error) {
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.DefinitionFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(ctx.Query("sort"), ctx.Query("direction"), port.DefaultDefinitionSort(), port.AllowedDefinitionSorts())
	if err != nil {
		return port.DefinitionFilter{}, searchProblem(err)
	}
	return port.DefinitionFilter{
		Query:  query,
		Sort:   sort,
		Status: domain.DefinitionStatus(ctx.Query("status")),
	}, nil
}

// issueRequest defines package data.
type issueRequest struct {
	DefinitionID  uuid.UUID         `json:"definition_id"`  // DefinitionID stores the definition i d value.
	TargetUserID  uuid.UUID         `json:"target_user_id"` // TargetUserID stores the target user i d value.
	TargetIPHash  string            `json:"target_ip_hash"` // TargetIPHash stores the target i p hash value.
	IssuerType    domain.IssuerType `json:"issuer_type"`    // IssuerType stores the issuer type value.
	IssuerKey     string            `json:"issuer_key"`     // IssuerKey stores the issuer key value.
	Reason        string            `json:"reason"`         // Reason stores the reason value.
	PrivateReason string            `json:"private_reason"` // PrivateReason stores the private reason value.
	StartsAt      *time.Time        `json:"starts_at"`      // StartsAt stores the starts at value.
	ExpiresAt     *time.Time        `json:"expires_at"`     // ExpiresAt stores the expires at value.
	Source        string            `json:"source"`         // Source stores the source value.
}

// issueCommand maps HTTP issue input to an application command.
func issueCommand(ctx *fiber.Ctx, actor uuid.UUID, request issueRequest) port.IssueCommand {
	startsAt := time.Time{}
	if request.StartsAt != nil {
		startsAt = *request.StartsAt
	}
	issuerUserID := &actor
	return port.IssueCommand{
		ActorUserID:    actor,
		DefinitionID:   request.DefinitionID,
		TargetUserID:   request.TargetUserID,
		TargetIPHash:   request.TargetIPHash,
		IssuerType:     request.IssuerType,
		IssuerUserID:   issuerUserID,
		IssuerKey:      request.IssuerKey,
		Reason:         request.Reason,
		PrivateReason:  request.PrivateReason,
		StartsAt:       startsAt,
		ExpiresAt:      request.ExpiresAt,
		Source:         request.Source,
		IdempotencyKey: ctx.Get(headers.IdempotencyKey),
	}
}
