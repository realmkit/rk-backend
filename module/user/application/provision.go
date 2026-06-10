package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/user/domain"
	"github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/api/auth"
	"github.com/niflaot/gamehub-go/pkg/api/principal"
	"github.com/niflaot/gamehub-go/pkg/identity"
)

// Provision resolves or creates the local user for identity.
func (service Service) Provision(ctx context.Context, external identity.ExternalIdentity, token auth.Token) (principal.Principal, error) {
	link, err := service.links.FindByIssuerSubject(ctx, external.Issuer, external.Subject)
	if err == nil {
		return service.provisionExisting(ctx, link, external, token)
	}
	if !errors.Is(err, port.ErrNotFound) {
		return principal.Principal{}, err
	}
	user, link, created, err := service.createProvisionedUser(ctx, external)
	if err != nil && !errors.Is(err, port.ErrConflict) {
		return principal.Principal{}, err
	}
	if created {
		if err := service.publishUserEvent(ctx, userProvisionedEvent, user, user.ID); err != nil {
			return principal.Principal{}, err
		}
		if err := service.publishIdentityEvent(ctx, identityLinkedEvent, link); err != nil {
			return principal.Principal{}, err
		}
	}
	link, err = service.links.FindByIssuerSubject(ctx, external.Issuer, external.Subject)
	if err != nil {
		return principal.Principal{}, err
	}
	return service.provisionExisting(ctx, link, external, token)
}

// DevelopmentPrincipal returns a principal for an existing local development user.
func (service Service) DevelopmentPrincipal(ctx context.Context, userID uuid.UUID) (principal.Principal, error) {
	user, err := service.users.FindByID(ctx, userID)
	if err != nil {
		return principal.Principal{}, err
	}
	if !user.CanAuthenticate() {
		return principal.Principal{}, disabledError()
	}
	return principal.Principal{UserID: user.ID, Issuer: "gamehub-development", SubjectHash: "dev:" + user.ID.String(), Scopes: []string{"development"}, DevelopmentBypass: true}, nil
}

// provisionExisting updates last-seen and claim cache for an existing identity.
func (service Service) provisionExisting(ctx context.Context, link domain.IdentityLink, external identity.ExternalIdentity, token auth.Token) (principal.Principal, error) {
	user, err := service.users.FindByID(ctx, link.UserID)
	if err != nil {
		return principal.Principal{}, err
	}
	if !user.CanAuthenticate() {
		return principal.Principal{}, disabledError()
	}
	now := service.clock()
	link.LastSeenAt = &now
	link.LastSyncedAt = &now
	link.ClaimsHash = external.RawClaimsHash
	if err := service.links.Touch(ctx, link); err != nil {
		return principal.Principal{}, err
	}
	if err := service.users.TouchLastSeen(ctx, user.ID); err != nil {
		return principal.Principal{}, err
	}
	if _, err := service.claims.Upsert(ctx, claimCacheFromExternal(user.ID, external, now)); err != nil {
		return principal.Principal{}, err
	}
	principal := service.principalFor(user, link, token, false)
	return principal, service.publishIdentityEvent(ctx, identityClaimRefreshedEvent, link)
}

// createProvisionedUser creates user, link, and claim cache in one transaction.
func (service Service) createProvisionedUser(
	ctx context.Context,
	external identity.ExternalIdentity,
) (domain.User, domain.IdentityLink, bool, error) {
	var createdUser domain.User
	var createdLink domain.IdentityLink
	err := service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		now := service.clock()
		user, err := service.users.Create(ctx, domain.User{ID: uuid.New(), Status: domain.StatusActive, FirstSeenAt: now, LastSeenAt: &now, Version: 1})
		if err != nil {
			return err
		}
		link := domain.IdentityLink{ID: uuid.New(), UserID: user.ID, Provider: service.provider, Issuer: external.Issuer, Subject: external.Subject, SubjectHash: identity.SubjectHash(external.Issuer, external.Subject), ClaimsHash: external.RawClaimsHash, LinkedAt: now, LastSeenAt: &now, LastSyncedAt: &now}
		if err := link.Validate(); err != nil {
			return err
		}
		if _, err := service.links.Create(ctx, link); err != nil {
			return err
		}
		if _, err := service.claims.Upsert(ctx, claimCacheFromExternal(user.ID, external, now)); err != nil {
			return err
		}
		createdUser = user
		createdLink = link
		return nil
	})
	return createdUser, createdLink, err == nil, err
}

// claimCacheFromExternal maps external identity to claim cache.
func claimCacheFromExternal(userID uuid.UUID, external identity.ExternalIdentity, syncedAt time.Time) domain.ClaimCache {
	return domain.ClaimCache{ID: uuid.New(), UserID: userID, Issuer: external.Issuer, Subject: external.Subject, Username: external.Username, Email: external.Email, EmailVerified: external.EmailVerified, DisplayName: external.DisplayName, PictureURL: external.PictureURL, PreferredLocale: external.PreferredLocale, ClaimsHash: external.RawClaimsHash, SyncedAt: syncedAt}
}
