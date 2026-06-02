package http

import (
	"encoding/json"
	"net/http"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/jsonapi"
	"github.com/fromforgesoftware/go-kit/search/query"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

// ServiceAccountController serves the admin CRUD for machine identities (JSON:API)
// plus the public client-credentials token endpoint (an SA's own login, plain JSON).
type ServiceAccountController struct {
	accounts app.ServiceAccountUsecase
	roles    app.RoleUsecase
	authz    app.AuthzUsecase
	auth     app.AuthUsecase
}

func NewServiceAccountController(accounts app.ServiceAccountUsecase, roles app.RoleUsecase, authz app.AuthzUsecase, auth app.AuthUsecase) kitrest.Controller {
	return &ServiceAccountController{accounts: accounts, roles: roles, authz: authz, auth: auth}
}

func (c *ServiceAccountController) gate(action string, h http.Handler) http.Handler {
	return guard(c.auth, c.authz, action, h)
}

func (c *ServiceAccountController) Routes(r kitrest.Router) {
	r.Method(http.MethodGet, "/api/admin/service-accounts", c.gate("service_accounts.read",
		kitrest.NewJsonApiListHandler(c.accounts, api.ServiceAccountToDTO)))
	r.Method(http.MethodGet, "/api/admin/service-accounts/{id}", c.gate("service_accounts.read",
		kitrest.NewJsonApiGetHandler(c.accounts, api.ServiceAccountToDTO, []query.ParseOpt{})))
	r.Method(http.MethodPost, "/api/admin/service-accounts", c.gate("service_accounts.write",
		http.HandlerFunc(c.create)))
	r.Method(http.MethodDelete, "/api/admin/service-accounts/{id}", c.gate("service_accounts.write",
		kitrest.NewJsonApiDeleteHandler(c.accounts, repository.DeleteTypeHard)))
	r.Method(http.MethodGet, "/api/admin/service-accounts/{id}/roles", c.gate("service_accounts.read",
		http.HandlerFunc(c.getRoles)))
	r.Method(http.MethodPut, "/api/admin/service-accounts/{id}/roles", c.gate("service_accounts.write",
		http.HandlerFunc(c.setRoles)))

	r.Post("/api/auth/service-accounts/token", http.HandlerFunc(c.token))
}

// create returns the one-time credentials (including the plaintext secret) as a
// JSON:API document — a CREATE that doesn't fit the resource encoder because the
// secret is shown exactly once and never re-readable.
func (c *ServiceAccountController) create(w http.ResponseWriter, r *http.Request) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.ServiceAccountCreateDTO](r)
	if err != nil {
		writeErr(w, err)
		return
	}
	creds, err := c.accounts.Create(r.Context(), body.Name())
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.api+json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = jsonapi.MarshalPayload(w, api.ServiceAccountCredentialsToDTO(creds))
}

func (c *ServiceAccountController) getRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.roles.RolesForServiceAccount(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeRolesJSONAPI(w, roles)
}

func (c *ServiceAccountController) setRoles(w http.ResponseWriter, r *http.Request) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.SetRolesDTO](r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := c.roles.SetServiceAccountRoles(r.Context(), app.SetServiceAccountRolesCommand{
		ServiceAccountID: r.PathValue("id"),
		Roles:            body.Roles(),
	}); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *ServiceAccountController) token(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "malformed body"})
		return
	}
	token, expiresIn, err := c.accounts.IssueToken(r.Context(), body.ClientID, body.ClientSecret)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"accessToken": token,
		"tokenType":   "Bearer",
		"expiresIn":   expiresIn,
	})
}
