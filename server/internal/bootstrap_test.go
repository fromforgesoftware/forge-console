package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/go-kit/monitoring/logger"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// stubAppRepo captures the apps ensureApps upserts so we can assert the parsed
// fields (notably the derived moduleUri).
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

func TestLoadAppsConfigParsesInstallAndEnable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "apps.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`install:
  - id: aegis
    bundle: "oci://ghcr.io/fromforgesoftware/aegis-console:0.1.0"
  - id: talos
    bundle: "oci://ghcr.io/fromforgesoftware/talos-console:0.1.0"
enable:
  - id: aegis
    apiBase: "http://aegis"
  - id: catalog
    apiBase: "http://catalog"
    name: "Catalog"
`), 0o600))

	cfg := loadAppsConfig(path)

	require.Len(t, cfg.Install, 2)
	require.Equal(t, "aegis", cfg.Install[0].ID)
	require.Equal(t, "oci://ghcr.io/fromforgesoftware/aegis-console:0.1.0", cfg.Install[0].Bundle)
	require.Equal(t, "talos", cfg.Install[1].ID)

	require.Len(t, cfg.Enable, 2)
	require.Equal(t, "aegis", cfg.Enable[0].ID)
	require.Equal(t, "http://aegis", cfg.Enable[0].APIBase)
	require.Equal(t, "catalog", cfg.Enable[1].ID)
	require.Equal(t, "Catalog", cfg.Enable[1].Name)
}

func TestLoadAppsConfigMissingFileIsEmpty(t *testing.T) {
	cfg := loadAppsConfig(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	require.Empty(t, cfg.Install)
	require.Empty(t, cfg.Enable)
}

func TestEnsureAppsDerivesModuleURI(t *testing.T) {
	cfg := bootstrapConfig{
		apps: appsConfig{
			// aegis is enabled AND installed → derived moduleUri.
			// catalog is enabled only (no install bundle) → empty moduleUri.
			Install: []installEntry{
				{ID: "aegis", Bundle: "oci://ghcr.io/fromforgesoftware/aegis-console:0.1.0"},
			},
			Enable: []enableEntry{
				{ID: "aegis", APIBase: "http://aegis"},
				{ID: "catalog", APIBase: "http://catalog", Name: "Catalog"},
			},
		},
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
	require.Equal(t, "aegis", aegis.Name(), "name defaults from id when not set")
	require.Equal(t, "http://aegis", aegis.AdminBaseURL())
	require.True(t, aegis.Enabled())
	require.Equal(t, "/public/plugins/aegis/module.js", aegis.ModuleURI(),
		"enabled+installed app gets a derived moduleUri")

	catalog := byslug["catalog"]
	require.NotNil(t, catalog)
	require.Equal(t, "Catalog", catalog.Name())
	require.Equal(t, "http://catalog", catalog.AdminBaseURL())
	require.Empty(t, catalog.ModuleURI(), "enabled-only app has empty moduleUri")
}

func TestEnsureAppsSkipsEmptyID(t *testing.T) {
	cfg := bootstrapConfig{
		apps: appsConfig{
			Enable: []enableEntry{
				{ID: "", APIBase: "http://nowhere"},
				{ID: "talos", APIBase: "http://talos"},
			},
		},
	}
	repo := &stubAppRepo{}

	err := ensureApps(context.Background(), cfg, repo, stubPerms{}, logger.New())
	require.NoError(t, err)
	require.Len(t, repo.upserted, 1)
	require.Equal(t, "talos", repo.upserted[0].Slug())
}
