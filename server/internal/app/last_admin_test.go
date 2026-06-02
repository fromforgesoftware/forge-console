package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/go-kit/application/repository"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

type stubUsers struct {
	patched bool
}

func (s *stubUsers) Get(context.Context, ...search.Option) (app.User, error) {
	return app.NewUser("admin@example.com", app.WithUserID("user-1")), nil
}
func (s *stubUsers) List(context.Context, ...search.Option) (resource.ListResponse[app.User], error) {
	return resource.NewEmptyListResponse[app.User](), nil
}
func (s *stubUsers) Create(_ context.Context, u app.User) (app.User, error) { return u, nil }
func (s *stubUsers) Patch(_ context.Context, _ ...repository.PatchOption) ([]app.User, error) {
	s.patched = true
	return []app.User{app.NewUser("admin@example.com", app.WithUserID("user-1"))}, nil
}

// adminRoles backs the SetUserRoles / Patch guard: one role granting "*.*"
// bound to the only enabled admin.
type adminRoles struct {
	stubRoles
}

func newAdminRoles(count int) *adminRoles {
	return &adminRoles{stubRoles{roles: []app.Role{role("admin", "*.*")}, adminCount: count}}
}

func TestPatchUser_DisableLastAdmin_Conflicts(t *testing.T) {
	uc := app.NewUserUsecase(&stubUsers{}, nil, newAdminRoles(1), app.NewAuthzUsecase(newAdminRoles(1)), nil)
	_, err := uc.Patch(context.Background(), app.PatchUserCommand{UserID: "user-1", Status: app.UserDisabled})
	require.Error(t, err)
	require.True(t, apierrors.Is(err, apierrors.CodeConflict))
}

func TestPatchUser_DisableWithSpareAdmin_OK(t *testing.T) {
	users := &stubUsers{}
	uc := app.NewUserUsecase(users, nil, newAdminRoles(2), app.NewAuthzUsecase(newAdminRoles(2)), nil)
	_, err := uc.Patch(context.Background(), app.PatchUserCommand{UserID: "user-1", Status: app.UserDisabled})
	require.NoError(t, err)
	require.True(t, users.patched)
}

func TestSetUserRoles_StripLastAdmin_Conflicts(t *testing.T) {
	// New role set is empty → would remove "*.*"; only one admin remains.
	roles := newAdminRoles(1)
	uc := app.NewUserUsecase(&stubUsers{}, nil, roles, app.NewAuthzUsecase(roles), nil)
	_, err := uc.SetRoles(context.Background(), app.SetUserRolesCommand{UserID: "user-1", Roles: nil})
	require.Error(t, err)
	require.True(t, apierrors.Is(err, apierrors.CodeConflict))
}
