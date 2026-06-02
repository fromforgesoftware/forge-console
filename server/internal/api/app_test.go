package api

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

func TestAppNavToDTOSurfacesModuleURI(t *testing.T) {
	a := app.NewApp("aegis",
		app.WithAppName("Aegis"),
		app.WithAppKind("aegis"),
		app.WithAppAdminBaseURL("http://forge-aegis"),
		app.WithAppModuleURI("http://forge-aegis/console/remoteEntry.js"),
		app.WithAppEnabled(true),
	)
	dto := AppNavToDTO(a)
	require.NotNil(t, dto)
	require.Equal(t, "aegis", dto.RSlug)
	require.Equal(t, "Aegis", dto.RName)
	require.Equal(t, "http://forge-aegis/console/remoteEntry.js", dto.RModuleURI)
}

func TestAppNavToDTOEmptyModuleURI(t *testing.T) {
	a := app.NewApp("catalog",
		app.WithAppName("Catalog"),
		app.WithAppAdminBaseURL("http://forge-catalog"),
		app.WithAppEnabled(true),
	)
	dto := AppNavToDTO(a)
	require.NotNil(t, dto)
	require.Empty(t, dto.RModuleURI)
}

func TestAppToDTOSurfacesModuleURI(t *testing.T) {
	a := app.NewApp("aegis",
		app.WithAppName("Aegis"),
		app.WithAppAdminBaseURL("http://forge-aegis"),
		app.WithAppModuleURI("http://forge-aegis/console/remoteEntry.js"),
		app.WithAppEnabled(true),
	)
	dto := AppToDTO(a)
	require.NotNil(t, dto)
	require.Equal(t, "http://forge-aegis", dto.RAdminBaseURL)
	require.Equal(t, "http://forge-aegis/console/remoteEntry.js", dto.RModuleURI)
}
