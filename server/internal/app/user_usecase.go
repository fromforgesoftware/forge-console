package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	"github.com/fromforgesoftware/go-kit/auth/password"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
)

// adminPermission is the wildcard grant whose holders must never drop below one.
const adminPermission = "*.*"

// CreateUserCommand creates a console administrator with an optional local password.
type CreateUserCommand struct {
	Email       string
	DisplayName string
	Password    string
}

// PatchUserCommand mutates a user's status (the only mutable field via admin).
type PatchUserCommand struct {
	UserID string
	Status UserStatus
}

// SetUserRolesCommand replaces the full set of roles bound to a user.
type SetUserRolesCommand struct {
	UserID string
	Roles  []string
}

// UserUsecase is the admin read/write surface over users plus the role-binding
// and status guards that protect the last admin.
type UserUsecase interface {
	repository.Getter[User]
	repository.Lister[User]

	Create(ctx context.Context, cmd CreateUserCommand) (User, error)
	Patch(ctx context.Context, cmd PatchUserCommand) (User, error)
	SetRoles(ctx context.Context, cmd SetUserRolesCommand) (User, error)
	RolesForUser(ctx context.Context, userID string) ([]Role, error)
}

type userUsecase struct {
	repository.Getter[User]
	repository.Lister[User]

	users  UserRepository
	creds  CredentialRepository
	roles  RoleRepository
	authz  AuthzUsecase
	hasher password.Hasher
}

func NewUserUsecase(users UserRepository, creds CredentialRepository, roles RoleRepository, authz AuthzUsecase, hasher password.Hasher) UserUsecase {
	return &userUsecase{
		Getter: usecase.NewGetter[User](users, ResourceTypeUser),
		Lister: usecase.NewLister[User](users),
		users:  users,
		creds:  creds,
		roles:  roles,
		authz:  authz,
		hasher: hasher,
	}
}

func (uc *userUsecase) Create(ctx context.Context, cmd CreateUserCommand) (User, error) {
	email := normalizeEmail(cmd.Email)
	if email == "" {
		return nil, apierrors.InvalidArgument("email is required")
	}
	u, err := uc.users.Create(ctx, NewUser(email,
		WithUserDisplayName(cmd.DisplayName),
		WithUserStatus(UserEnabled),
	))
	if err != nil {
		return nil, err
	}
	if cmd.Password != "" {
		hashed, err := uc.hasher.Hash(cmd.Password)
		if err != nil {
			return nil, err
		}
		if err := uc.creds.Set(ctx, u.ID(), hashed.Encoded); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func (uc *userUsecase) Patch(ctx context.Context, cmd PatchUserCommand) (User, error) {
	if cmd.Status == UserDisabled {
		if err := uc.guardLastAdmin(ctx, cmd.UserID); err != nil {
			return nil, err
		}
	}
	updated, err := uc.users.Patch(ctx,
		repository.PatchSearchOpts(search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", cmd.UserID))),
		repository.PatchField("status", string(cmd.Status)),
	)
	if err != nil {
		return nil, err
	}
	if len(updated) == 0 {
		return nil, apierrors.NotFound("user", cmd.UserID)
	}
	return updated[0], nil
}

func (uc *userUsecase) SetRoles(ctx context.Context, cmd SetUserRolesCommand) (User, error) {
	wasAdmin, err := uc.authz.IsAdmin(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if wasAdmin && !rolesGrantAdmin(ctx, uc.roles, cmd.Roles) {
		if err := uc.guardLastAdmin(ctx, cmd.UserID); err != nil {
			return nil, err
		}
	}
	if err := uc.roles.SetSubjectRoles(ctx, SubjectTypeUser, cmd.UserID, cmd.Roles); err != nil {
		return nil, err
	}
	return uc.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", cmd.UserID)))
}

func (uc *userUsecase) RolesForUser(ctx context.Context, userID string) ([]Role, error) {
	return uc.roles.RolesForSubject(ctx, SubjectTypeUser, userID)
}

// guardLastAdmin rejects an operation that would leave zero enabled users
// holding the "*.*" wildcard grant.
func (uc *userUsecase) guardLastAdmin(ctx context.Context, userID string) error {
	holdsAdmin, err := uc.authz.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if !holdsAdmin {
		return nil
	}
	count, err := uc.roles.CountEnabledUsersWithPermission(ctx, adminPermission)
	if err != nil {
		return err
	}
	if count <= 1 {
		return apierrors.Conflict("at least one admin must remain")
	}
	return nil
}

// rolesGrantAdmin reports whether the given role set grants the "*.*" wildcard.
func rolesGrantAdmin(ctx context.Context, repo RoleRepository, roleSlugs []string) bool {
	for _, slug := range roleSlugs {
		r, err := repo.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", slug)))
		if err != nil || r == nil {
			continue
		}
		for _, p := range r.Permissions() {
			if permissionMatches(p, adminPermission) {
				return true
			}
		}
	}
	return false
}
