package port

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
)

// TestAllowAllPolicyAllowsEveryOperation verifies the local development policy.
func TestAllowAllPolicyAllowsEveryOperation(t *testing.T) {
	policy := AllowAllPolicy{}
	actor := Actor{ID: uuid.New()}
	owner := OwnerRef{Type: domain.OwnerUser, ID: uuid.New()}

	if err := policy.CanManageDefinitions(context.Background(), actor); err != nil {
		t.Fatalf("CanManageDefinitions() error = %v", err)
	}
	if err := policy.CanReadOwnerMetadata(context.Background(), actor, owner); err != nil {
		t.Fatalf("CanReadOwnerMetadata() error = %v", err)
	}
	if err := policy.CanWriteOwnerMetadata(context.Background(), actor, owner); err != nil {
		t.Fatalf("CanWriteOwnerMetadata() error = %v", err)
	}
	if err := policy.CanManageMetaobjects(context.Background(), actor); err != nil {
		t.Fatalf("CanManageMetaobjects() error = %v", err)
	}
}
