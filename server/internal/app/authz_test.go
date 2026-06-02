package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// stubRoles is a hand stub for RoleRepository — only RolesForSubject and the
// last-admin count are exercised.
type stubRoles struct {
	roles      []app.Role
	adminCount int
}

func (s stubRoles) Get(context.Context, ...search.Option) (app.Role, error) {
	if len(s.roles) == 0 {
		return nil, nil
	}
	return s.roles[0], nil
}

func (s stubRoles) List(context.Context, ...search.Option) (resource.ListResponse[app.Role], error) {
	return resource.NewListResponse(s.roles, len(s.roles)), nil
}

func (s stubRoles) Create(_ context.Context, r app.Role) (app.Role, error) { return r, nil }

func (s stubRoles) Delete(context.Context, repository.DeleteType, ...search.Option) error { return nil }

func (s stubRoles) BindSubject(context.Context, app.SubjectType, string, string) error { return nil }

func (s stubRoles) SetSubjectRoles(context.Context, app.SubjectType, string, []string) error {
	return nil
}

func (s stubRoles) RolesForSubject(context.Context, app.SubjectType, string) ([]app.Role, error) {
	return s.roles, nil
}

func (s stubRoles) CountEnabledUsersWithPermission(context.Context, string) (int, error) {
	return s.adminCount, nil
}

func role(slug string, perms ...string) app.Role {
	return app.NewRole(slug, app.WithRolePermissions(perms))
}

func TestAuthz_Admin(t *testing.T) {
	uc := app.NewAuthzUsecase(stubRoles{roles: []app.Role{role("admin", "*.*")}})
	isAdmin, err := uc.IsAdmin(context.Background(), "user-1")
	require.NoError(t, err)
	require.True(t, isAdmin)
	ok, _ := uc.CanAccessApp(context.Background(), "user-1", "aegis")
	require.True(t, ok)
}

func TestAuthz_ScopedToApps(t *testing.T) {
	uc := app.NewAuthzUsecase(stubRoles{roles: []app.Role{role("hallmark-only", "app:hallmark.read")}})
	can, err := uc.CanAccessApp(context.Background(), "user-1", "hallmark")
	require.NoError(t, err)
	require.True(t, can)
	cannot, _ := uc.CanAccessApp(context.Background(), "user-1", "aegis")
	require.False(t, cannot)
}

func TestAuthz_NoRolesDeniesAll(t *testing.T) {
	uc := app.NewAuthzUsecase(stubRoles{})
	ok, err := uc.CanAccessApp(context.Background(), "user-1", "hallmark")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAuthz_EffectivePermissions(t *testing.T) {
	uc := app.NewAuthzUsecase(stubRoles{roles: []app.Role{
		role("a", "users.read", "roles.read"),
		role("b", "users.read", "apps.write"),
	}})
	perms, err := uc.EffectivePermissions(context.Background(), app.SubjectTypeUser, "user-1")
	require.NoError(t, err)
	require.Equal(t, []string{"apps.write", "roles.read", "users.read"}, perms)
}

func TestAuthz_Can(t *testing.T) {
	uc := app.NewAuthzUsecase(stubRoles{roles: []app.Role{role("viewer", "*.read")}})
	ok, err := uc.Can(context.Background(), app.SubjectTypeUser, "user-1", "users.read")
	require.NoError(t, err)
	require.True(t, ok)
	denied, _ := uc.Can(context.Background(), app.SubjectTypeUser, "user-1", "users.write")
	require.False(t, denied)
}
