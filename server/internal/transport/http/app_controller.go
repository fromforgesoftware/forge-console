package http

import (
	"net/http"

	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

// AppController exposes the app registry that drives the SPA nav,
// scoped to what the signed-in user may access. The admin base URL is
// intentionally not exposed — it's gateway-internal.
type AppController struct {
	apps  app.AppUsecase
	auth  app.AuthUsecase
	authz app.AuthzUsecase
}

func NewAppController(apps app.AppUsecase, auth app.AuthUsecase, authz app.AuthzUsecase) kitrest.Controller {
	return &AppController{apps: apps, auth: auth, authz: authz}
}

func (c *AppController) Routes(r kitrest.Router) {
	r.Get("/api/apps", http.HandlerFunc(c.list))
}

func (c *AppController) list(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	apps, err := c.apps.ListEnabled(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]app.App, 0, len(apps))
	for _, a := range apps {
		allowed, err := c.authz.CanAccessApp(r.Context(), u.ID(), a.Slug())
		if err != nil {
			writeErr(w, err)
			return
		}
		if allowed {
			out = append(out, a)
		}
	}
	writeManyJSONAPI(w, out, api.AppNavToDTO)
}
