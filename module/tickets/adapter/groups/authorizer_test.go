package groups

import (
	"context"
	"testing"

	"github.com/google/uuid"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	groupsport "github.com/niflaot/gamehub-go/module/groups/port"
)

// TestAuthorizerChecksTicketAndPunishmentPermissions verifies permission translation.
func TestAuthorizerChecksTicketAndPunishmentPermissions(t *testing.T) {
	checker := &checkerFake{allowed: map[groupsdomain.Permission]bool{
		groupsdomain.PermissionTicketsCreate:         true,
		groupsdomain.PermissionTicketsPerformActions: true,
		groupsdomain.PermissionPunishmentsRevoke:     true,
	}}
	authorizer := NewAuthorizer(checker)
	allowed, err := authorizer.CanCreate(context.Background(), uuid.New(), uuid.New())
	if err != nil || !allowed {
		t.Fatalf("CanCreate() = %v, %v; want allowed", allowed, err)
	}
	allowed, err = authorizer.CanRevokePunishmentFromAppeal(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != nil || !allowed {
		t.Fatalf("CanRevokePunishmentFromAppeal() = %v, %v; want allowed", allowed, err)
	}
	if got := checker.requests[1].ObjectType; got != groupsdomain.ObjectTicket {
		t.Fatalf("ticket object type = %s, want ticket", got)
	}
	if got := checker.requests[2].ObjectType; got != groupsdomain.ObjectPunishment {
		t.Fatalf("punishment object type = %s, want punishment", got)
	}
}

// TestAuthorizerDeniesWithoutChecker verifies fail-closed behavior.
func TestAuthorizerDeniesWithoutChecker(t *testing.T) {
	authorizer := NewAuthorizer(nil)
	allowed, err := authorizer.CanView(context.Background(), uuid.New(), uuid.New())
	if err != nil || allowed {
		t.Fatalf("CanView() = %v, %v; want denied without error", allowed, err)
	}
}

type checkerFake struct {
	allowed  map[groupsdomain.Permission]bool
	requests []groupsport.CheckRequest
}

func (checker *checkerFake) Check(
	_ context.Context,
	request groupsport.CheckRequest,
) (groupsport.Decision, error) {
	checker.requests = append(checker.requests, request)
	return groupsport.Decision{Allowed: checker.allowed[request.Permission]}, nil
}
