package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// AppDTO is the jsonapi representation of a managed app registry entry.
type AppDTO struct {
	resource.RestDTO

	RSlug         string `jsonapi:"attr,slug"`
	RName         string `jsonapi:"attr,name"`
	RKind         string `jsonapi:"attr,kind"`
	RAdminBaseURL string `jsonapi:"attr,adminBaseURL"`
	REnabled      bool   `jsonapi:"attr,enabled"`
}

func AppToDTO(a app.App) *AppDTO {
	if a == nil {
		return nil
	}
	dto := &AppDTO{
		RestDTO:       resource.ToRestDTO(a),
		RSlug:         a.Slug(),
		RName:         a.Name(),
		RKind:         a.Kind(),
		RAdminBaseURL: a.AdminBaseURL(),
		REnabled:      a.Enabled(),
	}
	dto.RType = app.ResourceTypeApp
	return dto
}

// AppNavDTO is the nav-scoped view of an app — slug/name/kind only. The admin
// base URL is gateway-internal and intentionally omitted (unlike AppDTO).
type AppNavDTO struct {
	resource.RestDTO

	RSlug string `jsonapi:"attr,slug"`
	RName string `jsonapi:"attr,name"`
	RKind string `jsonapi:"attr,kind"`
}

func AppNavToDTO(a app.App) *AppNavDTO {
	if a == nil {
		return nil
	}
	dto := &AppNavDTO{
		RestDTO: resource.ToRestDTO(a),
		RSlug:   a.Slug(),
		RName:   a.Name(),
		RKind:   a.Kind(),
	}
	dto.RType = app.ResourceTypeApp
	return dto
}
