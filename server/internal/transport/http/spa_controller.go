package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"
)

// SPAController serves the built Foundry SPA: real files when present, else
// index.html so the client router handles deep links. It registers "/" and a
// "/{path...}" catch-all — /api/* and /readyz are more specific patterns and
// take precedence. When FOUNDRY_STATIC_DIR is unset (API-only dev) it registers
// nothing.
type SPAController struct {
	dir string
}

func NewSPAController() kitrest.Controller {
	return &SPAController{dir: os.Getenv("FOUNDRY_STATIC_DIR")}
}

func (c *SPAController) Routes(r kitrest.Router) {
	if c.dir == "" {
		return
	}
	// "/{path...}" matches "/" and every deep link; /api/* and /readyz are more
	// specific patterns and take precedence.
	r.Get("/{path...}", http.HandlerFunc(c.serve))
}

func (c *SPAController) serve(w http.ResponseWriter, r *http.Request) {
	if rel := strings.TrimPrefix(r.URL.Path, "/"); rel != "" {
		p := filepath.Join(c.dir, filepath.Clean("/"+rel))
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			http.ServeFile(w, r, p)
			return
		}
	}
	// SPA shell for "/" and any client-router deep link. Served directly (not
	// via ServeFile) to avoid its /index.html→/ canonical redirect, and never
	// cached so a redeploy's new bundle is picked up.
	b, err := os.ReadFile(filepath.Join(c.dir, "index.html"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}
