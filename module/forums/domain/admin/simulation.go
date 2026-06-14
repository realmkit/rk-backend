package admin

import (
	"strings"

	"github.com/google/uuid"
)

// ForumPermissionSimulationRequest requests an explainable permission decision.
type ForumPermissionSimulationRequest struct {
	// ActorUserID is the selected actor; nil means anonymous.
	ActorUserID uuid.UUID `json:"actor_user_id,omitempty"`

	// Permission is the permission name to evaluate.
	Permission string `json:"permission"`

	// ObjectType is the object type to explain.
	ObjectType string `json:"object_type"`

	// ObjectID is the object identifier to explain.
	ObjectID uuid.UUID `json:"object_id,omitempty"`
}

// Normalize returns a normalized simulation request.
func (request ForumPermissionSimulationRequest) Normalize(
	forumID uuid.UUID,
) ForumPermissionSimulationRequest {
	request.Permission = strings.TrimSpace(request.Permission)
	request.ObjectType = strings.TrimSpace(request.ObjectType)
	if request.ObjectType == "" {
		request.ObjectType = "forum"
	}
	if request.ObjectID == uuid.Nil {
		request.ObjectID = forumID
	}
	return request
}

// Validate validates a simulation request.
func (request ForumPermissionSimulationRequest) Validate() error {
	var violations []Violation
	if request.Permission == "" {
		violations = AppendViolation(violations, "permission", "is required")
	}
	if request.ObjectType == "" {
		violations = AppendViolation(violations, "object_type", "is required")
	}
	if request.ObjectID == uuid.Nil {
		violations = AppendViolation(violations, "object_id", "is required")
	}
	return NewValidationError(violations)
}

// ForumPermissionSimulationResult explains a forum permission decision.
type ForumPermissionSimulationResult struct {
	// Allowed reports whether the actor is allowed.
	Allowed bool `json:"allowed"`

	// Reason explains the decision.
	Reason string `json:"reason"`

	// Permission is the evaluated permission.
	Permission string `json:"permission"`

	// ObjectType is the evaluated object type.
	ObjectType string `json:"object_type"`

	// ObjectID is the evaluated object identifier.
	ObjectID uuid.UUID `json:"object_id"`

	// MatchedAction is the action grant that allowed the request.
	MatchedAction string `json:"matched_action,omitempty"`

	// CheckedActions are the action grants considered.
	CheckedActions []string `json:"checked_actions"`
}
