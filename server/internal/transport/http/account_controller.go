package http

import (
	"net/http"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

// AccountController is the signed-in user's self-service surface: edit own
// profile, change own password, read/write own settings. Cookie-authed,
// JSON:API on both sides (profile → `users`, settings → `user-settings`).
type AccountController struct {
	account app.AccountUsecase
	auth    app.AuthUsecase
	authz   app.AuthzUsecase
	roles   app.RoleRepository
}

func NewAccountController(account app.AccountUsecase, auth app.AuthUsecase, authz app.AuthzUsecase, roles app.RoleRepository) kitrest.Controller {
	return &AccountController{account: account, auth: auth, authz: authz, roles: roles}
}

func (c *AccountController) Routes(r kitrest.Router) {
	r.Patch("/api/users/me", http.HandlerFunc(c.updateProfile))
	r.Put("/api/users/me/password", http.HandlerFunc(c.changePassword))
	r.Get("/api/users/me/settings", http.HandlerFunc(c.getSettings))
	r.Put("/api/users/me/settings", http.HandlerFunc(c.updateSettings))
}

func (c *AccountController) updateProfile(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.MeProfilePatchDTO](r)
	if err != nil {
		writeErr(w, apierrors.InvalidArgument("malformed body"))
		return
	}
	updated, err := c.account.UpdateProfile(r.Context(), u.ID(), body.DisplayName())
	if err != nil {
		writeErr(w, err)
		return
	}
	settings, err := c.account.GetSettings(r.Context(), u.ID())
	if err != nil {
		writeErr(w, err)
		return
	}
	roleSlugs, perms, err := userRolesAndPermissions(r.Context(), c.authz, c.roles, u.ID())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOneJSONAPI(w, http.StatusOK, api.AuthUserToDTO(updated, settings, roleSlugs, perms))
}

func (c *AccountController) changePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.PasswordChangeDTO](r)
	if err != nil {
		writeErr(w, apierrors.InvalidArgument("malformed body"))
		return
	}
	if err := c.account.ChangePassword(r.Context(), u.ID(), body.CurrentPassword(), body.NewPassword()); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *AccountController) getSettings(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	s, err := c.account.GetSettings(r.Context(), u.ID())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOneJSONAPI(w, http.StatusOK, api.UserSettingsToDTO(u.ID(), s))
}

func (c *AccountController) updateSettings(w http.ResponseWriter, r *http.Request) {
	u, ok := resolveUser(w, r, c.auth)
	if !ok {
		return
	}
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.UserSettingsDTO](r)
	if err != nil {
		writeErr(w, apierrors.InvalidArgument("malformed body"))
		return
	}
	s, err := c.account.UpdateSettings(r.Context(), app.UserSettings{UserID: u.ID(), Theme: body.Theme()})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOneJSONAPI(w, http.StatusOK, api.UserSettingsToDTO(u.ID(), s))
}
