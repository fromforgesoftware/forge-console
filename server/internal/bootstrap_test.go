package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/go-kit/monitoring/logger"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// stubAppRepo captures the apps ensureApps upserts so we can assert the parsed
// fields (notably the optional moduleUri 4th field).
type stubAppRepo struct {
	upserted []app.App
}

func (s *stubAppRepo) Get(context.Context, ...search.Option) (app.App, error) { return nil, nil }
func (s *stubAppRepo) List(context.Context, ...search.Option) (resource.ListResponse[app.App], error) {
	return resource.NewEmptyListResponse[app.App](), nil
}
func (s *stubAppRepo) Upsert(_ context.Context, a app.App) error {
	s.upserted = append(s.upserted, a)
	return nil
}

// stubPerms swallows the per-app permission seeding ensureApps performs.
type stubPerms struct{}

func (stubPerms) List(context.Context, ...search.Option) (resource.ListResponse[app.Permission], error) {
	return resource.NewEmptyListResponse[app.Permission](), nil
}
func (stubPerms) Upsert(context.Context, app.Permission) error { return nil }

func TestEnsureAppsParsesModuleURI(t *testing.T) {
	cfg := bootstrapConfig{
		// aegis has the optional 4th field (moduleUri); catalog is a 3-field
		// entry (no console remote) and must still parse.
		apps: "aegis=Aegis=http://forge-aegis=http://forge-aegis/console/remoteEntry.js," +
			"catalog=Catalog=http://forge-catalog",
	}
	repo := &stubAppRepo{}

	err := ensureApps(context.Background(), cfg, repo, stubPerms{}, logger.New())
	require.NoError(t, err)
	require.Len(t, repo.upserted, 2)

	byslug := map[string]app.App{}
	for _, a := range repo.upserted {
		byslug[a.Slug()] = a
	}

	aegis := byslug["aegis"]
	require.NotNil(t, aegis)
	require.Equal(t, "Aegis", aegis.Name())
	require.Equal(t, "http://forge-aegis", aegis.AdminBaseURL())
	require.Equal(t, "http://forge-aegis/console/remoteEntry.js", aegis.ModuleURI())

	catalog := byslug["catalog"]
	require.NotNil(t, catalog)
	require.Equal(t, "http://forge-catalog", catalog.AdminBaseURL())
	require.Empty(t, catalog.ModuleURI(), "3-field entry must have empty moduleUri")
}

func TestEnsureAppsSkipsMalformedEntries(t *testing.T) {
	cfg := bootstrapConfig{
		// Fewer than 3 fields is malformed and skipped; the valid entry still
		// registers.
		apps: "broken=OnlyTwo,talos=Talos=http://forge-talos=http://forge-talos/console/remoteEntry.js",
	}
	repo := &stubAppRepo{}

	err := ensureApps(context.Background(), cfg, repo, stubPerms{}, logger.New())
	require.NoError(t, err)
	require.Len(t, repo.upserted, 1)
	require.Equal(t, "talos", repo.upserted[0].Slug())
	require.Equal(t, "http://forge-talos/console/remoteEntry.js", repo.upserted[0].ModuleURI())
}
