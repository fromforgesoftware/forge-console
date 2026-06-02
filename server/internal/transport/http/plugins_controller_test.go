package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"
)

func TestPluginsControllerServesModule(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "aegis"), 0o755))
	const body = "System.register([], function(){});"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "aegis", "module.js"), []byte(body), 0o644))

	t.Setenv("FORGE_PLUGINS_DIR", dir)
	h := kitrest.BuildHandler(NewPluginsController())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/public/plugins/aegis/module.js", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, body, rec.Body.String())
	require.Contains(t, rec.Header().Get("Content-Type"), "javascript",
		"module.js must be served with a JS content-type")
}

func TestPluginsControllerMissingFile404(t *testing.T) {
	t.Setenv("FORGE_PLUGINS_DIR", t.TempDir())
	h := kitrest.BuildHandler(NewPluginsController())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/public/plugins/aegis/module.js", nil))

	require.Equal(t, http.StatusNotFound, rec.Code)
}
