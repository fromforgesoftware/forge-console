package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
)

// UpsertRoleCommand creates or replaces a CUSTOM role and its permission set.
type UpsertRoleCommand struct {
	Slug        string
	Name        string
	Permissions []string
}

// SetServiceAccountRolesCommand replaces the roles bound to a service account.
type SetServiceAccountRolesCommand struct {
	ServiceAccountID string
	Roles            []string
}

// RoleUsecase is the admin read/write surface over roles. SYSTEM roles are
// immutable: create/update and delete reject them.
type RoleUsecase interface {
	repository.Getter[Role]
	repository.Lister[Role]

	Upsert(ctx context.Context, cmd UpsertRoleCommand) (Role, error)
	Delete(ctx context.Context, slug string) error
	RolesForServiceAccount(ctx context.Context, saID string) ([]Role, error)
	SetServiceAccountRoles(ctx context.Context, cmd SetServiceAccountRolesCommand) error
}

type roleUsecase struct {
	repository.Getter[Role]
	repository.Lister[Role]

	roles RoleRepository
}

func NewRoleUsecase(roles RoleRepository) RoleUsecase {
	return &roleUsecase{
		Getter: usecase.NewGetter[Role](roles, ResourceTypeRole),
		Lister: usecase.NewLister[Role](roles),
		roles:  roles,
	}
}

func (uc *roleUsecase) Upsert(ctx context.Context, cmd UpsertRoleCommand) (Role, error) {
	if cmd.Slug == "" {
		return nil, apierrors.InvalidArgument("slug is required")
	}
	if existing, err := uc.roles.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", cmd.Slug))); err == nil && existing != nil {
		if existing.Kind() == RoleSystem {
			return nil, apierrors.Forbidden("system roles are immutable")
		}
	} else if err != nil && !apierrors.Is(err, apierrors.CodeNotFound) {
		return nil, err
	}
	return uc.roles.Create(ctx, NewRole(cmd.Slug,
		WithRoleName(cmd.Name),
		WithRoleKind(RoleCustom),
		WithRolePermissions(cmd.Permissions),
	))
}

func (uc *roleUsecase) Delete(ctx context.Context, slug string) error {
	existing, err := uc.roles.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", slug)))
	if err != nil {
		return err
	}
	if existing == nil {
		return apierrors.NotFound("role", slug)
	}
	if existing.Kind() == RoleSystem {
		return apierrors.Forbidden("system roles are immutable")
	}
	return uc.roles.Delete(ctx, repository.DeleteTypeHard, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", slug)))
}

func (uc *roleUsecase) RolesForServiceAccount(ctx context.Context, saID string) ([]Role, error) {
	return uc.roles.RolesForSubject(ctx, SubjectTypeServiceAccount, saID)
}

func (uc *roleUsecase) SetServiceAccountRoles(ctx context.Context, cmd SetServiceAccountRolesCommand) error {
	return uc.roles.SetSubjectRoles(ctx, SubjectTypeServiceAccount, cmd.ServiceAccountID, cmd.Roles)
}
