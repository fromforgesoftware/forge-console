package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// PermissionDTO is the jsonapi representation of a permission catalog entry.
type PermissionDTO struct {
	resource.RestDTO

	RResourceType string `jsonapi:"attr,resourceType"`
	RVerb         string `jsonapi:"attr,verb"`
	RDescription  string `jsonapi:"attr,description,omitempty"`
}

func PermissionToDTO(p app.Permission) *PermissionDTO {
	if p == nil {
		return nil
	}
	dto := &PermissionDTO{
		RestDTO:       resource.ToRestDTO(p),
		RResourceType: p.ResourceType(),
		RVerb:         p.Verb(),
		RDescription:  p.Description(),
	}
	dto.RType = app.ResourceTypePermission
	return dto
}
