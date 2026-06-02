package http

import (
	"context"
	"net/http"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

// AuthController serves user auth: local-password login, logout, and the
// current-user lookup. Login is modeled as a `sessions` resource (POST creates
// it + sets an httpOnly cookie); the response and `me` are the `users` resource.
// JSON:API on both sides — login/logout stay hand-written only because they must
// set the session cookie (the kit command handler doesn't expose the writer).
type AuthController struct {
	auth    app.AuthUsecase
	account app.AccountUsecase
	authz   app.AuthzUsecase
	roles   app.RoleRepository
}

func NewAuthController(auth app.AuthUsecase, account app.AccountUsecase, authz app.AuthzUsecase, roles app.RoleRepository) kitrest.Controller {
	return &AuthController{auth: auth, account: account, authz: authz, roles: roles}
}

func (c *AuthController) Routes(r kitrest.Router) {
	r.Post("/api/auth/login", http.HandlerFunc(c.login))
	r.Post("/api/auth/logout", http.HandlerFunc(c.logout))
	r.Get("/api/users/me", http.HandlerFunc(c.me))
}

// hydrate builds the current-user resource, loading settings + effective
// permissions/roles so one GET feeds the SPA (drives theme + permission nav).
func (c *AuthController) hydrate(ctx context.Context, u app.User) (*api.AuthUserDTO, error) {
	s, err := c.account.GetSettings(ctx, u.ID())
	if err != nil {
		return nil, err
	}
	roles, perms, err := userRolesAndPermissions(ctx, c.authz, c.roles, u.ID())
	if err != nil {
		return nil, err
	}
	return api.AuthUserToDTO(u, s, roles, perms), nil
}

func (c *AuthController) login(w http.ResponseWriter, r *http.Request) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.SessionCreateDTO](r)
	if err != nil {
		writeErr(w, apierrors.InvalidArgument("malformed body"))
		return
	}
	s, err := c.auth.Login(r.Context(), body.REmail, body.RPassword)
	if err != nil {
		writeErr(w, err)
		return
	}
	setSessionCookie(w, r, s.ID, int(sessionMaxAgeSeconds))
	u, err := c.auth.Authenticate(r.Context(), s.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	dto, err := c.hydrate(r.Context(), u)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOneJSONAPI(w, http.StatusCreated, dto)
}

func (c *AuthController) logout(w http.ResponseWriter, r *http.Request) {
	if ck, err := r.Cookie(sessionCookie); err == nil {
		_ = c.auth.Logout(r.Context(), ck.Value)
	}
	setSessionCookie(w, r, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

func (c *AuthController) me(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	dto, err := c.hydrate(r.Context(), u)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOneJSONAPI(w, http.StatusOK, dto)
}

const sessionMaxAgeSeconds = 12 * 60 * 60
