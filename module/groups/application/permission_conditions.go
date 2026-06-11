package application

import (
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
)

// conditionsMatch reports whether all policy conditions pass.
func (service Service) conditionsMatch(
	request port.CheckRequest,
	conditions []domain.PolicyCondition,
) (bool, []domain.PolicyCondition, error) {
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
