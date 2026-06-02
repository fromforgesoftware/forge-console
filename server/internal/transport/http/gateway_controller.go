package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/fromforgesoftware/go-kit/auth/jwt"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// GatewayController reverse-proxies user requests to a managed app's
// admin API: ALL /api/proxy/{app}/{path...} → {app.adminBaseURL}/{path}.
// It requires an authenticated user (session cookie) and resolves the
// target from the app registry, so the upstream location is configuration,
// not a same-cluster assumption. When FORGE_GATEWAY_SECRET is set it also mints
// a short HMAC token the apps verify, so they only accept gateway traffic.
type GatewayController struct {
	apps   app.AppUsecase
	auth   app.AuthUsecase
	authz  app.AuthzUsecase
	issuer jwt.Issuer // nil when no gateway secret is configured
}

func NewGatewayController(apps app.AppUsecase, auth app.AuthUsecase, authz app.AuthzUsecase) kitrest.Controller {
	c := &GatewayController{apps: apps, auth: auth, authz: authz}
	if secret := os.Getenv("FORGE_GATEWAY_SECRET"); secret != "" {
		if iss, err := jwt.NewHMACIssuer(secret); err == nil {
			c.issuer = iss
		}
	}
	return c
}

// saValidator returns the issuer as a JWT validator (it is also an
// IssuerValidator) so the gateway can accept service-account bearer tokens.
func (c *GatewayController) saValidator() jwt.Validator {
	if c.issuer == nil {
		return nil
	}
	if v, ok := c.issuer.(jwt.Validator); ok {
		return v
	}
	return nil
}

func (c *GatewayController) Routes(r kitrest.Router) {
	for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		r.Method(m, "/api/proxy/{app}/{path...}", http.HandlerFunc(c.proxy))
	}
}

func (c *GatewayController) proxy(w http.ResponseWriter, r *http.Request) {
	subj, ok := resolveSubject(w, r, c.auth, c.saValidator())
	if !ok {
		return
	}

	slug := r.PathValue("app")
	a, err := c.apps.Get(r.Context(), slug)
	if err != nil {
		writeErr(w, err)
		return
	}
	allowed, err := c.canAccess(r, subj, slug)
	if err != nil {
		writeErr(w, err)
		return
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	base, err := url.Parse(a.AdminBaseURL())
	if err != nil || base.Host == "" {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "app has no valid admin URL"})
		return
	}

	upstreamPath := strings.TrimRight(base.Path, "/") + "/" + r.PathValue("path")
	rawQuery := r.URL.RawQuery
	gatewayToken := c.mintToken(r, subj)
	(&httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(base)
			pr.Out.URL.Path = upstreamPath
			pr.Out.URL.RawQuery = rawQuery
			pr.Out.Host = base.Host
			// Don't leak the Foundry user session to the app.
			pr.Out.Header.Del("Cookie")
			if gatewayToken != "" {
				pr.Out.Header.Set("Authorization", "Bearer "+gatewayToken)
			}
		},
	}).ServeHTTP(w, r)
}

// canAccess resolves whether the subject may reach the app, branching on its
// kind (users and service accounts both authorize via role bindings).
func (c *GatewayController) canAccess(r *http.Request, subj Subject, slug string) (bool, error) {
	if subj.Kind == SubjectServiceAccount {
		return c.authz.CanServiceAccountAccessApp(r.Context(), subj.ID, slug)
	}
	return c.authz.CanAccessApp(r.Context(), subj.ID, slug)
}

// mintToken issues the short-lived HMAC token apps verify (empty when no
// gateway secret is configured — apps then stay open).
func (c *GatewayController) mintToken(r *http.Request, subj Subject) string {
	if c.issuer == nil {
		return ""
	}
	id, err := uuid.Parse(subj.ID)
	if err != nil {
		id = uuid.Nil
	}
	tok, err := c.issuer.Issue(r.Context(), id, subj.ID)
	if err != nil {
		return ""
	}
	return tok
}
