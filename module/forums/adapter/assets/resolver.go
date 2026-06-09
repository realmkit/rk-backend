package assets

import (
	"context"
	"errors"

	"github.com/google/uuid"
	assetport "github.com/niflaot/gamehub-go/module/assets/port"
	forumsport "github.com/niflaot/gamehub-go/module/forums/port"
)

// Resolver validates forum attachment references through the assets service.
type Resolver struct {
	service assetport.Service
}

// NewResolver creates an asset resolver.
func NewResolver(service assetport.Service) Resolver {
	return Resolver{service: service}
}

// AssetExists reports whether an asset exists.
func (resolver Resolver) AssetExists(ctx context.Context, id uuid.UUID) (bool, error) {
	if resolver.service == nil {
		return false, nil
	}
	if _, err := resolver.service.Get(ctx, id); err != nil {
		if errors.Is(err, assetport.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Ensure Resolver implements the forum asset resolver.
var _ forumsport.AssetResolver = Resolver{}
