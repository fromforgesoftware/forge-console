package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// RoleDTO is the jsonapi representation of a role and its permission set.
type RoleDTO struct {
	resource.RestDTO

	RSlug        string   `jsonapi:"attr,slug"`
	RName        string   `jsonapi:"attr,name"`
	RKind        string   `jsonapi:"attr,kind"`
	RPermissions []string `jsonapi:"attr,permissions"`
}

func RoleToDTO(r app.Role) *RoleDTO {
	if r == nil {
		return nil
	}
	perms := r.Permissions()
	if perms == nil {
		perms = []string{}
	}
	dto := &RoleDTO{
		RestDTO:      resource.ToRestDTO(r),
		RSlug:        r.Slug(),
		RName:        r.Name(),
		RKind:        string(r.Kind()),
		RPermissions: perms,
	}
	dto.RType = app.ResourceTypeRole
	return dto
}
