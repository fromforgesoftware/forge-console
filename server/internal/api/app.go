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
	RModuleURI    string `jsonapi:"attr,moduleUri"`
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
		RModuleURI:    a.ModuleURI(),
		REnabled:      a.Enabled(),
	}
	dto.RType = app.ResourceTypeApp
	return dto
}

// AppNavDTO is the nav-scoped view of an app. The admin base URL is
// gateway-internal and intentionally omitted (unlike AppDTO). ModuleURI is a
// browser-reachable Module-Federation remote (remoteEntry.js) the console
// loads at runtime; it's safe to expose and empty for apps without a console
// remote.
type AppNavDTO struct {
	resource.RestDTO

	RSlug      string `jsonapi:"attr,slug"`
	RName      string `jsonapi:"attr,name"`
	RKind      string `jsonapi:"attr,kind"`
	RModuleURI string `jsonapi:"attr,moduleUri"`
}

func AppNavToDTO(a app.App) *AppNavDTO {
	if a == nil {
		return nil
	}
	dto := &AppNavDTO{
		RestDTO:    resource.ToRestDTO(a),
		RSlug:      a.Slug(),
		RName:      a.Name(),
		RKind:      a.Kind(),
		RModuleURI: a.ModuleURI(),
	}
	dto.RType = app.ResourceTypeApp
	return dto
}
