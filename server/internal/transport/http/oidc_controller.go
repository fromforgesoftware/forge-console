package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/fromforgesoftware/go-kit/auth/oidc"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

const oidcFlowCookie = "foundry_oidc_flow"

// OIDCController handles external (pluggable) user sign-in. The provider is
// just config — Aegis, Google, GitHub, etc. The external identity is mapped to
// an EXISTING user by email; there is no just-in-time provisioning
// (users only, invite/seed to add one).
type OIDCController struct {
	providers app.OIDCProviders
	users     app.UserRepository
	auth      app.AuthUsecase
}

func NewOIDCController(providers app.OIDCProviders, users app.UserRepository, auth app.AuthUsecase) kitrest.Controller {
	return &OIDCController{providers: providers, users: users, auth: auth}
}

func (c *OIDCController) Routes(r kitrest.Router) {
	r.Get("/api/auth/providers", http.HandlerFunc(c.list))
	r.Get("/api/auth/oidc/{provider}/start", http.HandlerFunc(c.start))
	r.Get("/api/auth/oidc/{provider}/callback", http.HandlerFunc(c.callback))
	r.Get("/api/auth/logout", http.HandlerFunc(c.logout))
}

// logout clears the Foundry session and, when an OIDC provider is configured,
// bounces through its end_session endpoint (RP-initiated logout) so the IdP
// session is terminated too — otherwise SSO would silently sign the user back
// in. It then lands on the login screen. Browser navigation (GET), not XHR.
func (c *OIDCController) logout(w http.ResponseWriter, r *http.Request) {
	if ck, err := r.Cookie(sessionCookie); err == nil {
		_ = c.auth.Logout(r.Context(), ck.Value)
	}
	setSessionCookie(w, r, "", -1)
	for _, p := range c.providers.List() {
		if p.Issuer != "" {
			login := c.baseURL(r) + "/login"
			http.Redirect(w, r, p.Issuer+"/logout?post_logout_redirect_uri="+neturl.QueryEscape(login), http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (c *OIDCController) baseURL(r *http.Request) string {
	scheme := "http"
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		scheme = p
	} else if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func (c *OIDCController) list(w http.ResponseWriter, _ *http.Request) {
	provs := c.providers.List()
	writeManyJSONAPI(w, provs, func(p app.OIDCProvider) *api.AuthProviderDTO {
		return api.AuthProviderToDTO(p.ID, p.Name)
	})
}

type oidcFlow struct {
	Provider string `json:"provider"`
	Verifier string `json:"verifier"`
	State    string `json:"state"`
}

func (c *OIDCController) start(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("provider")
	prov, ok := c.providers.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	verifier, err := oidc.NewVerifier()
	if err != nil {
		c.fail(w, r, "oidc_init")
		return
	}
	state, _ := oidc.NewVerifier()
	authURL, err := prov.Client.AuthCodeURL(r.Context(), c.callbackURL(r, id), state, oidc.Challenge(verifier))
	if err != nil {
		c.fail(w, r, "oidc_unavailable")
		return
	}
	c.setFlowCookie(w, r, oidcFlow{Provider: id, Verifier: verifier, State: state})
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (c *OIDCController) callback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("provider")
	prov, ok := c.providers.Get(id)
	if !ok {
		c.fail(w, r, "unknown_provider")
		return
	}
	flow, err := c.readFlowCookie(r)
	c.clearFlowCookie(w)
	q := r.URL.Query()
	if err != nil || flow.Provider != id || flow.State == "" || q.Get("state") != flow.State {
		c.fail(w, r, "oidc_state")
		return
	}
	code := q.Get("code")
	if code == "" {
		c.fail(w, r, "oidc_no_code")
		return
	}
	toks, err := prov.Client.Exchange(r.Context(), c.callbackURL(r, id), code, flow.Verifier)
	if err != nil {
		c.fail(w, r, "oidc_exchange")
		return
	}
	claims, err := prov.Client.VerifyIDToken(r.Context(), toks.IDToken)
	if err != nil || claims.Email == "" {
		c.fail(w, r, "oidc_verify")
		return
	}
	u, err := app.GetUserByEmail(r.Context(), c.users, strings.ToLower(strings.TrimSpace(claims.Email)))
	if err != nil {
		// Users only — the identity is valid but isn't a user here.
		c.fail(w, r, "not_a_user")
		return
	}
	sess, err := c.auth.StartSession(r.Context(), u.ID())
	if err != nil {
		c.fail(w, r, "session")
		return
	}
	setSessionCookie(w, r, sess.ID, sessionMaxAgeSeconds)
	http.Redirect(w, r, "/", http.StatusFound)
}

// fail bounces back to the SPA login with an error code (browser navigation).
func (c *OIDCController) fail(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, "/login?error="+code, http.StatusFound)
}

func (c *OIDCController) callbackURL(r *http.Request, provider string) string {
	scheme := "http"
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		scheme = p
	} else if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/api/auth/oidc/" + provider + "/callback"
}

func (c *OIDCController) setFlowCookie(w http.ResponseWriter, r *http.Request, f oidcFlow) {
	b, _ := json.Marshal(f)
	http.SetCookie(w, &http.Cookie{
		Name: oidcFlowCookie, Value: base64.RawURLEncoding.EncodeToString(b),
		Path: "/api/auth/oidc", HttpOnly: true, Secure: secureRequest(r), SameSite: http.SameSiteLaxMode, MaxAge: 600,
	})
}

func (c *OIDCController) readFlowCookie(r *http.Request) (oidcFlow, error) {
	var f oidcFlow
	ck, err := r.Cookie(oidcFlowCookie)
	if err != nil {
		return f, err
	}
	raw, err := base64.RawURLEncoding.DecodeString(ck.Value)
	if err != nil {
		return f, err
	}
	return f, json.Unmarshal(raw, &f)
}

func (c *OIDCController) clearFlowCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: oidcFlowCookie, Value: "", Path: "/api/auth/oidc", MaxAge: -1})
}
