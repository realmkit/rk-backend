package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/transaction"
)

// Dependencies contains user service dependencies.
type Dependencies struct {
	// Users stores local users.
	Users port.UserRepository

	// Links stores identity links.
	Links port.IdentityLinkRepository

	// Claims stores provider claim caches.
	Claims port.ClaimCacheRepository

	// Transactions runs provisioning atomically.
	Transactions transaction.Runner

	// Provider is the configured identity provider preset.
	Provider string

	// Events publishes user lifecycle events.
	Events emitter.Publisher
}

// Service manages users and identity provisioning.
type Service struct {
	users        port.UserRepository         // users stores the users value.
	links        port.IdentityLinkRepository // links stores the links value.
	claims       port.ClaimCacheRepository   // claims stores the claims value.
	transactions transaction.Runner          // transactions stores the transactions value.
	provider     string                      // provider stores the provider value.
	clock        func() time.Time            // clock stores the clock value.
	events       emitter.Publisher           // events stores the events value.
}

// NewService creates a user service.
func NewService(dependencies Dependencies) Service {
	return Service{
		users:        dependencies.Users,
		links:        dependencies.Links,
		claims:       dependencies.Claims,
		transactions: dependencies.Transactions,
		provider:     dependencies.Provider,
		clock:        func() time.Time { return time.Now().UTC() },
		events:       dependencies.Events,
	}
}

// Get returns one user.
func (service Service) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return service.users.FindByID(ctx, id)
}

// Current returns the current user aggregate.
func (service Service) Current(ctx context.Context, userID uuid.UUID) (port.CurrentUser, error) {
	user, err := service.users.FindByID(ctx, userID)
	if err != nil {
		return port.CurrentUser{}, err
	}
	claims, err := service.claims.FindByUserID(ctx, userID)
	if err != nil && !errors.Is(err, port.ErrNotFound) {
		return port.CurrentUser{}, err
	}
	result := port.CurrentUser{User: user}
	if err == nil {
		result.Claims = &claims
	}
	return result, nil
}

// List returns matching users.
func (service Service) List(
	ctx context.Context,
	filter port.UserFilter,
	page pagination.Page,
) (pagination.Result[port.UserSummary], error) {
	return service.users.List(ctx, filter, page)
}

// FindSummariesByIDs returns display summaries for the requested local users.
func (service Service) FindSummariesByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) (map[uuid.UUID]port.UserSummary, error) {
	return service.users.FindSummariesByIDs(ctx, ids)
}

// UpdateCurrent updates local settings for the current user.
func (service Service) UpdateCurrent(ctx context.Context, command port.UpdateCurrentCommand) (domain.User, error) {
	user, err := service.users.FindByID(ctx, command.UserID)
	if err != nil {
		return domain.User{}, err
	}
	user.AvatarAssetID = command.AvatarAssetID
	if err := user.Validate(); err != nil {
		return domain.User{}, err
	}
	updated, err := service.users.Update(ctx, user, command.ExpectedVersion)
	if err != nil {
		return domain.User{}, err
	}
	return updated, service.publishUserEvent(ctx, userUpdatedEvent, updated, command.UserID)
}

// principalFor returns an authenticated principal for user and identity.
func (service Service) principalFor(user domain.User, link domain.IdentityLink, token auth.Token, development bool) principal.Principal {
	return principal.Principal{
		UserID:            user.ID,
		Issuer:            link.Issuer,
		Subject:           link.Subject,
		SubjectHash:       link.SubjectHash,
		Audience:          token.Audience,
		Scopes:            token.Scopes,
		DevelopmentBypass: development,
	}
}

// disabledError returns an auth-aware disabled user error.
func disabledError() error {
	return fmt.Errorf("%w: %w", auth.ErrDisabledUser, port.ErrDisabled)
}
