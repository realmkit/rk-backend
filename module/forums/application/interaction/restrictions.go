package interaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

func (service Service) requireUnrestricted(
	ctx context.Context,
	actorUserID uuid.UUID,
	actionKey string,
) error {
	if service.restrictions == nil || actorUserID == uuid.Nil {
		return nil
	}
	restricted, err := service.restrictions.Restricted(ctx, actorUserID, actionKey)
	if err != nil {
		return err
	}
	if restricted {
		return port.ErrForbidden
	}
	return nil
}
