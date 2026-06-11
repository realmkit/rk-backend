package application

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Check returns an authorization decision.
func (service Service) Check(ctx context.Context, request port.CheckRequest) (port.Decision, error) {
	rules, objectType, err := service.permissionRules(ctx, request.Permission)
	if err != nil {
		return port.Decision{Allowed: false, Reason: "unknown_permission"}, err
	}
	if len(rules) == 0 {
		return port.Decision{Allowed: false, Reason: "unknown_permission"}, port.ErrUnknownPermission
	}
	if objectType != request.ObjectType {
		return port.Decision{Allowed: false, Reason: "object_type_mismatch"}, nil
	}
	if request.ObjectID == uuid.Nil {
		return port.Decision{Allowed: false, Reason: "missing_identifier"}, nil
	}
	return service.checkRelations(ctx, request, rules)
}

// permissionRules returns configured rules or built-in fallback rules.
func (service Service) permissionRules(
	ctx context.Context,
	permission domain.Permission,
) ([]domain.PermissionRule, domain.ObjectType, error) {
	if service.policies != nil {
		definition, err := service.policies.FindDefinition(ctx, permission)
		if err == nil {
			if !definition.Enabled {
				return nil, definition.ObjectType, nil
			}
			rules, err := service.policies.ListRules(ctx, permission)
			if err != nil {
				return nil, "", err
			}
			rules = enabledRules(rules)
			return rules, definition.ObjectType, nil
		}
		if err != nil && !errors.Is(err, port.ErrNotFound) {
			return nil, "", err
		}
	}
	rule, ok := staticPermissionRules[permission]
	if !ok {
		return nil, "", port.ErrUnknownPermission
	}
	return staticRules(permission, rule), rule.objectType, nil
}

// enabledRules filters and orders active policy rules.
func enabledRules(rules []domain.PermissionRule) []domain.PermissionRule {
	result := make([]domain.PermissionRule, 0, len(rules))
	for _, rule := range rules {
		if rule.Enabled {
			result = append(result, rule)
		}
	}
	sort.SliceStable(result, func(left int, right int) bool {
		return result[left].Priority < result[right].Priority
	})
	return result
}

// staticRules converts static relation lists to policy rules.
func staticRules(permission domain.Permission, rule staticPermissionRule) []domain.PermissionRule {
	rules := make([]domain.PermissionRule, 0, len(rule.relations))
	for index, relation := range rule.relations {
		rules = append(
			rules,
			domain.PermissionRule{
				ID:         uuid.New(),
				Permission: permission,
				ObjectType: rule.objectType,
				Relation:   relation,
				Priority:   index,
				Enabled:    true,
			},
		)
	}
	return rules
}

// checkRelations checks one permission's allowed policy rules.
func (service Service) checkRelations(
	ctx context.Context,
	request port.CheckRequest,
	rules []domain.PermissionRule,
) (port.Decision, error) {
	tuples, err := service.tuples.List(
		ctx,
		port.TupleFilter{ObjectType: request.ObjectType, ObjectID: request.ObjectID},
		pagination.Page{Limit: 100},
	)
	if err != nil {
		return port.Decision{}, err
	}
	var failed []domain.PolicyCondition
	for _, tuple := range tuples.Items {
		ok, err := service.subjectMatches(ctx, request.ActorUserID, tuple)
		if err != nil {
			return port.Decision{}, err
		}
		if !ok {
			continue
		}
		for _, rule := range matchingRules(tuple.Relation, rules) {
			conditionsOK, failedConditions, err := service.conditionsMatch(request, rule.Conditions)
			if err != nil {
				return port.Decision{}, err
			}
			if conditionsOK {
				return port.Decision{
					Allowed:           true,
					Reason:            "matched_relation",
					MatchedRelation:   tuple.Relation,
					MatchedConditions: rule.Conditions,
				}, nil
			}
			failed = append(failed, failedConditions...)
		}
	}
	if len(failed) > 0 {
		return port.Decision{Allowed: false, Reason: "conditions_failed", FailedConditions: failed}, nil
	}
	return port.Decision{Allowed: false, Reason: "no_matching_relation"}, nil
}

// matchingRules returns policy rules for relation.
func matchingRules(relation domain.Relation, rules []domain.PermissionRule) []domain.PermissionRule {
	matches := []domain.PermissionRule{}
	for _, rule := range rules {
		if rule.Relation == relation {
			matches = append(matches, rule)
		}
	}
	return matches
}

// subjectMatches reports whether actor matches tuple subject.
func (service Service) subjectMatches(ctx context.Context, actorUserID uuid.UUID, tuple domain.RelationTuple) (bool, error) {
	switch tuple.SubjectType {
	case domain.SubjectPublic:
		return tuple.SubjectID == domain.PublicSubjectID(), nil
	case domain.SubjectAuthenticated:
		return actorUserID != uuid.Nil && tuple.SubjectID == domain.AuthenticatedSubjectID(), nil
	case domain.SubjectUser:
		return actorUserID != uuid.Nil && tuple.SubjectID == actorUserID, nil
	case domain.SubjectGroup:
		if actorUserID == uuid.Nil || tuple.SubjectRelation != domain.RelationMember {
			return false, nil
		}
		return service.activeGroupMember(ctx, tuple.SubjectID, actorUserID)
	default:
		return false, nil
	}
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
