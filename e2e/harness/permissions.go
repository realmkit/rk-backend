package harness

import (
	"context"

	groupsport "github.com/realmkit/rk-backend/module/groups/port"
)

// AllowChecker permits group-backed permission guards in e2e fixtures.
type AllowChecker struct{}

// Check returns an allowed permission decision.
func (AllowChecker) Check(context.Context, groupsport.CheckRequest) (groupsport.Decision, error) {
	return groupsport.Decision{Allowed: true, Reason: "e2e_allowed"}, nil
}
