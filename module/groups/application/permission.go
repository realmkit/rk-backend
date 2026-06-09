package application

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// staticPermissionRule defines built-in fallback policy rules.
type staticPermissionRule struct {
	objectType domain.ObjectType
	relations  []domain.Relation
}

// staticPermissionRules maps permissions to built-in fallback requirements.
var staticPermissionRules = map[domain.Permission]staticPermissionRule{
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
	if request.ActorUserID == uuid.Nil || request.ObjectID == uuid.Nil {
		return port.Decision{Allowed: false, Reason: "missing_identifier"}, nil
	}
	return service.checkRelations(ctx, request, rules)
}

// permissionRules returns configured rules or built-in fallback rules.
func (service Service) permissionRules(ctx context.Context, permission domain.Permission) ([]domain.PermissionRule, domain.ObjectType, error) {
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
		rules = append(rules, domain.PermissionRule{ID: uuid.New(), Permission: permission, ObjectType: rule.objectType, Relation: relation, Priority: index, Enabled: true})
	}
	return rules
}

// checkRelations checks one permission's allowed policy rules.
func (service Service) checkRelations(ctx context.Context, request port.CheckRequest, rules []domain.PermissionRule) (port.Decision, error) {
	tuples, err := service.tuples.List(ctx, port.TupleFilter{ObjectType: request.ObjectType, ObjectID: request.ObjectID}, pagination.Page{Limit: 100})
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
				return port.Decision{Allowed: true, Reason: "matched_relation", MatchedRelation: tuple.Relation, MatchedConditions: rule.Conditions}, nil
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

// conditionsMatch reports whether all policy conditions pass.
func (service Service) conditionsMatch(request port.CheckRequest, conditions []domain.PolicyCondition) (bool, []domain.PolicyCondition, error) {
	var failed []domain.PolicyCondition
	for _, condition := range conditions {
		ok, err := service.conditionMatches(request, condition)
		if err != nil {
			return false, nil, err
		}
		if !ok {
			failed = append(failed, condition)
		}
	}
	return len(failed) == 0, failed, nil
}

// conditionMatches reports whether one policy condition passes.
func (service Service) conditionMatches(request port.CheckRequest, condition domain.PolicyCondition) (bool, error) {
	value, found := request.Context[condition.Field]
	switch condition.Type {
	case domain.ConditionEquals:
		return found && stringValue(value) == condition.Value, nil
	case domain.ConditionIn:
		return found && slices.Contains(condition.Values, stringValue(value)), nil
	case domain.ConditionFieldEqualsActor, domain.ConditionAssignedToActor:
		return found && uuidValue(value) == request.ActorUserID, nil
	case domain.ConditionFieldNotEqualsActor:
		return !found || uuidValue(value) != request.ActorUserID, nil
	case domain.ConditionIsUnset:
		return !found || isEmpty(value), nil
	case domain.ConditionWithinDuration:
		return service.timestampWithinDuration(value, condition.Duration)
	case domain.ConditionOlderThan:
		return service.timestampOlderThan(value, condition.Duration)
	default:
		return false, nil
	}
}

// timestampWithinDuration reports whether value is inside the duration window.
func (service Service) timestampWithinDuration(value any, duration string) (bool, error) {
	instant, ok := timeValue(value)
	if !ok {
		return false, nil
	}
	period, err := time.ParseDuration(duration)
	if err != nil {
		return false, err
	}
	now := service.clock()
	return !instant.After(now) && now.Sub(instant) <= period, nil
}

// timestampOlderThan reports whether value is older than the duration.
func (service Service) timestampOlderThan(value any, duration string) (bool, error) {
	instant, ok := timeValue(value)
	if !ok {
		return false, nil
	}
	period, err := time.ParseDuration(duration)
	if err != nil {
		return false, err
	}
	return service.clock().Sub(instant) >= period, nil
}

// stringValue normalizes scalar context values to string.
func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case domain.ObjectType:
		return string(typed)
	case domain.Relation:
		return string(typed)
	case uuid.UUID:
		return typed.String()
	default:
		return ""
	}
}

// uuidValue normalizes context values to UUID.
func uuidValue(value any) uuid.UUID {
	switch typed := value.(type) {
	case uuid.UUID:
		return typed
	case string:
		parsed, err := uuid.Parse(typed)
		if err != nil {
			return uuid.Nil
		}
		return parsed
	default:
		return uuid.Nil
	}
}

// timeValue normalizes context values to time.
func timeValue(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), true
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, typed)
		if err != nil {
			return time.Time{}, false
		}
		return parsed.UTC(), true
	default:
		return time.Time{}, false
	}
}

// isEmpty reports whether a context value should be considered unset.
func isEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case uuid.UUID:
		return typed == uuid.Nil
	default:
		return false
	}
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
