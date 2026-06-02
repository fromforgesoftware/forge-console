// Package api holds Foundry's JSON:API DTOs for the admin resource surface.
package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// UserDTO is the jsonapi representation of a console administrator.
type UserDTO struct {
	resource.RestDTO

	REmail       string `jsonapi:"attr,email"`
	RDisplayName string `jsonapi:"attr,displayName,omitempty"`
	RStatus      string `jsonapi:"attr,status"`
}

func UserToDTO(u app.User) *UserDTO {
	if u == nil {
		return nil
	}
	dto := &UserDTO{
		RestDTO:      resource.ToRestDTO(u),
		REmail:       u.Email(),
		RDisplayName: u.DisplayName(),
		RStatus:      string(u.Status()),
	}
	dto.RType = app.ResourceTypeUser
	return dto
}
