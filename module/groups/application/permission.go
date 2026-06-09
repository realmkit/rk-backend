package application

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// permissionRule defines relations that allow a permission.
type permissionRule struct {
	objectType domain.ObjectType
	relations  []domain.Relation
}

// permissionRules maps permissions to relation requirements.
var permissionRules = map[domain.Permission]permissionRule{
	"groups.read":          {objectType: domain.ObjectGroup, relations: []domain.Relation{domain.RelationViewer, domain.RelationManager, domain.RelationMember, domain.RelationOwner}},
	"groups.update":        {objectType: domain.ObjectGroup, relations: []domain.Relation{domain.RelationManager, domain.RelationOwner}},
	"groups.delete":        {objectType: domain.ObjectGroup, relations: []domain.Relation{domain.RelationOwner}},
	"groups.assign_member": {objectType: domain.ObjectGroup, relations: []domain.Relation{domain.RelationManager, domain.RelationOwner}},
	"groups.read_members":  {objectType: domain.ObjectGroup, relations: []domain.Relation{domain.RelationViewer, domain.RelationManager, domain.RelationMember, domain.RelationOwner}},
	"assets.view":          {objectType: domain.ObjectAsset, relations: []domain.Relation{domain.RelationViewer, domain.RelationOwner}},
	"assets.update":        {objectType: domain.ObjectAsset, relations: []domain.Relation{domain.RelationEditor, domain.RelationOwner}},
	"metadata.write_user":  {objectType: domain.ObjectUser, relations: []domain.Relation{domain.RelationSelf, domain.RelationManager}},
}

// Check returns an authorization decision.
func (service Service) Check(ctx context.Context, request port.CheckRequest) (port.Decision, error) {
	rule, ok := permissionRules[request.Permission]
	if !ok {
		return port.Decision{Allowed: false, Reason: "unknown_permission"}, port.ErrUnknownPermission
	}
	if rule.objectType != request.ObjectType {
		return port.Decision{Allowed: false, Reason: "object_type_mismatch"}, nil
	}
	if request.ActorUserID == uuid.Nil || request.ObjectID == uuid.Nil {
		return port.Decision{Allowed: false, Reason: "missing_identifier"}, nil
	}
	return service.checkRelations(ctx, request, rule.relations)
}

// checkRelations checks one permission's allowed relations.
func (service Service) checkRelations(ctx context.Context, request port.CheckRequest, relations []domain.Relation) (port.Decision, error) {
	tuples, err := service.tuples.List(ctx, port.TupleFilter{ObjectType: request.ObjectType, ObjectID: request.ObjectID}, pagination.Page{Limit: 100})
	if err != nil {
		return port.Decision{}, err
	}
	for _, tuple := range tuples.Items {
		if !slices.Contains(relations, tuple.Relation) {
			continue
		}
		ok, err := service.subjectMatches(ctx, request.ActorUserID, tuple)
		if err != nil {
			return port.Decision{}, err
		}
		if ok {
			return port.Decision{Allowed: true, Reason: "matched_relation", MatchedRelation: tuple.Relation}, nil
		}
	}
	return port.Decision{Allowed: false, Reason: "no_matching_relation"}, nil
}

// subjectMatches reports whether actor matches tuple subject.
func (service Service) subjectMatches(ctx context.Context, actorUserID uuid.UUID, tuple domain.RelationTuple) (bool, error) {
	if tuple.SubjectType == domain.SubjectUser {
		return tuple.SubjectID == actorUserID, nil
	}
	if tuple.SubjectType == domain.SubjectGroup && tuple.SubjectRelation == domain.RelationMember {
		return service.activeGroupMember(ctx, tuple.SubjectID, actorUserID)
	}
	return false, nil
}

// activeGroupMember reports whether user is active in enabled group.
func (service Service) activeGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (bool, error) {
	group, err := service.groups.FindByID(ctx, groupID)
	if errors.Is(err, port.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !group.GrantsPermissions() {
		return false, nil
	}
	membership, err := service.memberships.Find(ctx, groupID, userID)
	if errors.Is(err, port.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return membership.ActiveAt(service.clock()), nil
}

// Ensure Service implements service contracts.
var _ port.GroupService = Service{}

// Ensure Service implements membership contracts.
var _ port.MembershipService = Service{}

// Ensure Service implements checker contracts.
var _ port.Checker = Service{}

// Ensure Service implements tuple contracts.
var _ port.TupleService = Service{}
