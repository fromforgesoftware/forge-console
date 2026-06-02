package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAppCarriesModuleURI(t *testing.T) {
	a := NewApp("aegis",
		WithAppName("Aegis"),
		WithAppKind("aegis"),
		WithAppAdminBaseURL("http://forge-aegis"),
		WithAppModuleURI("http://forge-aegis/console/remoteEntry.js"),
		WithAppEnabled(true),
	)
	require.Equal(t, "aegis", a.Slug())
	require.Equal(t, "http://forge-aegis", a.AdminBaseURL())
	require.Equal(t, "http://forge-aegis/console/remoteEntry.js", a.ModuleURI())
}

func TestNewAppModuleURIDefaultsEmpty(t *testing.T) {
	a := NewApp("catalog",
		WithAppName("Catalog"),
		WithAppAdminBaseURL("http://forge-catalog"),
		WithAppEnabled(true),
	)
	require.Empty(t, a.ModuleURI())
}
