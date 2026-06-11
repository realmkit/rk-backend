package domain

import (
	"github.com/google/uuid"
	adminmodel "github.com/realmkit/rk-backend/module/forums/domain/admin"
	structuremodel "github.com/realmkit/rk-backend/module/forums/domain/structure"
)

// Forum is a discussion board, container, or utility link.
type Forum = structuremodel.Forum

// ForumSettings contains admin-editable forum behavior settings.
type ForumSettings = structuremodel.ForumSettings

// ForumPermissionGrant grants one forum relation to one subject.
type ForumPermissionGrant = adminmodel.ForumPermissionGrant

// ForumPermissionSettings contains forum relation grants.
type ForumPermissionSettings = adminmodel.ForumPermissionSettings

// ForumPermissionSimulationRequest requests an explainable forum permission decision.
type ForumPermissionSimulationRequest = adminmodel.ForumPermissionSimulationRequest

// ForumPermissionSimulationResult explains a forum permission decision.
type ForumPermissionSimulationResult = adminmodel.ForumPermissionSimulationResult

// PublicPermissionSubjectID returns the public grant subject identifier.
func PublicPermissionSubjectID() uuid.UUID {
	return adminmodel.PublicPermissionSubjectID()
}

// AuthenticatedPermissionSubjectID returns the authenticated grant subject identifier.
func AuthenticatedPermissionSubjectID() uuid.UUID {
	return adminmodel.AuthenticatedPermissionSubjectID()
}
