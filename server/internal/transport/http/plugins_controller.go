package http

import (
	"net/http"
	"os"

	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"
)

// pluginsRoutePrefix is the browser-reachable base for installed console plugin
// assets. The forge-apps init-container unpacks each install bundle (a SystemJS
// module.js + assets) into <pluginsDir>/<id>/, and this route serves them at
// /public/plugins/<id>/module.js — the Grafana-equivalent of a plugin's
// module.js. moduleUri in GET /apps is derived from this prefix.
const pluginsRoutePrefix = "/public/plugins"

// PluginsController serves installed console plugin assets read-only from a
// shared volume the init-container populates. When FORGE_PLUGINS_DIR is unset
// it defaults to /var/lib/forge/plugins; the route is always registered (an
// empty dir simply 404s until a bundle is unpacked).
type PluginsController struct {
	dir string
}

func NewPluginsController() kitrest.Controller {
	dir := os.Getenv("FORGE_PLUGINS_DIR")
	if dir == "" {
		dir = "/var/lib/forge/plugins"
	}
	return &PluginsController{dir: dir}
}

func (c *PluginsController) Routes(r kitrest.Router) {
	// Method-scoped GET (not Mount): a methodless "/public/plugins/" pattern
	// conflicts with the SPA's "GET /{path...}" catch-all in Go's ServeMux
	// (each is more specific in a different dimension → registration panics).
	// "GET /public/plugins/{path...}" is strictly more specific, so both
	// coexist. StripPrefix keeps file-server paths relative to the plugins
	// dir (e.g. /aegis/module.js); http.FileServer sets the JS content-type
	// by extension and only ever reads (no writes).
	fs := http.StripPrefix(pluginsRoutePrefix, http.FileServer(http.Dir(c.dir)))
	r.Get(pluginsRoutePrefix+"/{path...}", http.HandlerFunc(fs.ServeHTTP))
}
