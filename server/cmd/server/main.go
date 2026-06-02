// Command server boots the Foundry control-plane backend: user auth
// (local password now; pluggable OIDC later), an app registry, and an
// admin-API gateway, served over the kit's REST gateway.
package main

import (
	"github.com/fromforgesoftware/go-kit/app"
	"github.com/fromforgesoftware/go-kit/openapi"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb/gormpg"

	"github.com/fromforgesoftware/forge/server/internal"
)

func main() {
	app.Run(
		app.WithName("foundry"),
		app.WithVersion(internal.Version),
		app.WithOpenAPI(
			openapi.SpecTitle("Foundry"),
			openapi.SpecVersion(internal.Version),
			openapi.SpecDescription("Forge user control plane."),
		),
		gormpg.FxModule(),
		internal.FxModule(),
	)
}
