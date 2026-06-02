package app

import (
	"context"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	"github.com/fromforgesoftware/go-kit/resource"
)

const ResourceTypePermission resource.Type = "permissions"

// Permission is a catalog entry naming an action with the grammar
// "<resourceType>.<verb>" (e.g. "users.read", "app:aegis.write"). The id is
// the full pattern.
type Permission interface {
	resource.Resource
	ResourceType() string
	Verb() string
	Description() string
}

type permission struct {
	resource.Resource

	resourceType string
	verb         string
	description  string
}

type PermissionOption func(*permission)

func WithPermissionResourceType(rt string) PermissionOption {
	return func(p *permission) { p.resourceType = rt }
}

func WithPermissionVerb(verb string) PermissionOption {
	return func(p *permission) { p.verb = verb }
}

func WithPermissionDescription(d string) PermissionOption {
	return func(p *permission) { p.description = d }
}

// NewPermission builds a Permission catalog entry keyed by id.
func NewPermission(id string, opts ...PermissionOption) Permission {
	p := &permission{
		Resource: resource.New(resource.WithType(ResourceTypePermission), resource.WithID(id)),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *permission) ResourceType() string { return p.resourceType }
func (p *permission) Verb() string         { return p.verb }
func (p *permission) Description() string  { return p.description }

// PermissionRepository persists the permission catalog the role picker lists.
type PermissionRepository interface {
	repository.Lister[Permission]
	Upsert(ctx context.Context, p Permission) error
}

// PermissionUsecase is the read surface over the permission catalog.
type PermissionUsecase interface {
	repository.Lister[Permission]
}

func NewPermissionUsecase(perms PermissionRepository) PermissionUsecase {
	return usecase.NewLister[Permission](perms)
}
