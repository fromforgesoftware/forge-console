package app

import (
	"context"
	"strings"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/auth/password"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
)

const minPasswordLen = 8

// AccountUsecase is the signed-in user's self-service surface: edit own
// profile, change own password, read/write own settings.
type AccountUsecase interface {
	UpdateProfile(ctx context.Context, userID, displayName string) (User, error)
	ChangePassword(ctx context.Context, userID, current, next string) error
	GetSettings(ctx context.Context, userID string) (UserSettings, error)
	UpdateSettings(ctx context.Context, s UserSettings) (UserSettings, error)
}

type accountUsecase struct {
	users    UserRepository
	creds    CredentialRepository
	settings SettingsRepository
	hasher   password.Hasher
}

func NewAccountUsecase(users UserRepository, creds CredentialRepository, settings SettingsRepository, hasher password.Hasher) AccountUsecase {
	return &accountUsecase{users: users, creds: creds, settings: settings, hasher: hasher}
}

func (uc *accountUsecase) UpdateProfile(ctx context.Context, userID, displayName string) (User, error) {
	updated, err := uc.users.Patch(ctx,
		repository.PatchSearchOpts(search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", userID))),
		repository.PatchField("displayName", strings.TrimSpace(displayName)),
	)
	if err != nil {
		return nil, err
	}
	if len(updated) == 0 {
		return nil, apierrors.NotFound("user", userID)
	}
	return updated[0], nil
}

func (uc *accountUsecase) ChangePassword(ctx context.Context, userID, current, next string) error {
	if len(next) < minPasswordLen {
		return apierrors.InvalidArgument("password must be at least 8 characters")
	}
	hash, err := uc.creds.HashFor(ctx, userID)
	if err != nil {
		if apierrors.Is(err, apierrors.CodeNotFound) {
			// External-OIDC-only users have no local credential to change.
			return apierrors.InvalidArgument("no local password set for this account")
		}
		return err
	}
	ok, err := uc.hasher.Verify(current, hash)
	if err != nil || !ok {
		return apierrors.InvalidArgument("current password is incorrect")
	}
	hashed, err := uc.hasher.Hash(next)
	if err != nil {
		return err
	}
	return uc.creds.Set(ctx, userID, hashed.Encoded)
}

func (uc *accountUsecase) GetSettings(ctx context.Context, userID string) (UserSettings, error) {
	return uc.settings.Get(ctx, userID)
}

func (uc *accountUsecase) UpdateSettings(ctx context.Context, s UserSettings) (UserSettings, error) {
	if err := uc.settings.Upsert(ctx, s); err != nil {
		return UserSettings{}, err
	}
	return uc.settings.Get(ctx, s.UserID)
}
