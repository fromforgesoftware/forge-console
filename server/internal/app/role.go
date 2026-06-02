package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/resource"
)

const ResourceTypeRole resource.Type = "roles"

// RoleKind separates seeded SYSTEM roles from operator-authored CUSTOM ones.
type RoleKind string

const (
	RoleSystem RoleKind = "SYSTEM"
	RoleCustom RoleKind = "CUSTOM"
)

// SubjectType identifies the kind of identity a role is bound to.
type SubjectType string

const (
	SubjectTypeUser           SubjectType = "USER"
	SubjectTypeServiceAccount SubjectType = "SERVICE_ACCOUNT"
)

// Role grants a set of permission patterns to its bound subjects. The slug is
// the resource id.
type Role interface {
	resource.Resource
	Slug() string
	Name() string
	Kind() RoleKind
	Permissions() []string
}

type role struct {
	resource.Resource

	name        string
	kind        RoleKind
	permissions []string
}

type RoleOption func(*role)

func WithRoleName(name string) RoleOption {
	return func(r *role) { r.name = name }
}

func WithRoleKind(kind RoleKind) RoleOption {
	return func(r *role) { r.kind = kind }
}

func WithRolePermissions(permissions []string) RoleOption {
	return func(r *role) { r.permissions = permissions }
}

// NewRole builds a Role aggregate keyed by slug.
func NewRole(slug string, opts ...RoleOption) Role {
	r := &role{
		Resource: resource.New(resource.WithType(ResourceTypeRole), resource.WithID(slug)),
		kind:     RoleCustom,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *role) Slug() string          { return r.ID() }
func (r *role) Name() string          { return r.name }
func (r *role) Kind() RoleKind        { return r.kind }
func (r *role) Permissions() []string { return r.permissions }

// RoleRepository persists roles, their permissions, and subject bindings.
type RoleRepository interface {
	repository.Getter[Role]
	repository.Lister[Role]
	repository.Creator[Role]
	repository.Deleter

	// BindSubject binds a single role to a subject (idempotent).
	BindSubject(ctx context.Context, subjectType SubjectType, subjectID, roleSlug string) error
	// SetSubjectRoles replaces the full set of roles bound to a subject.
	SetSubjectRoles(ctx context.Context, subjectType SubjectType, subjectID string, roleSlugs []string) error
	// RolesForSubject returns the roles bound to a subject, each with its permissions.
	RolesForSubject(ctx context.Context, subjectType SubjectType, subjectID string) ([]Role, error)
	// CountEnabledUsersWithPermission counts distinct enabled users holding a
	// permission, for the last-admin guard.
	CountEnabledUsersWithPermission(ctx context.Context, permission string) (int, error)
}
