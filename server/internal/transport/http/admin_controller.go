package http

import (
	"net/http"

	"github.com/fromforgesoftware/go-kit/search/query"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

// AdminController is the control-plane console: manage users, roles,
// permissions, and the app registry. The resource CRUD is JSON:API via the kit
// handlers; every route is gated on the matching "<resourceType>.<verb>"
// permission. The role-binding sub-resources are actions, not CRUD, so they
// ride JSON:API command handlers.
type AdminController struct {
	users app.UserUsecase
	roles app.RoleUsecase
	perms app.PermissionUsecase
	apps  app.AppAdminUsecase
	authz app.AuthzUsecase
	auth  app.AuthUsecase
}

func NewAdminController(users app.UserUsecase, roles app.RoleUsecase, perms app.PermissionUsecase, apps app.AppAdminUsecase, authz app.AuthzUsecase, auth app.AuthUsecase) kitrest.Controller {
	return &AdminController{users: users, roles: roles, perms: perms, apps: apps, authz: authz, auth: auth}
}

func (c *AdminController) gate(action string, h http.Handler) http.Handler {
	return guard(c.auth, c.authz, action, h)
}

func (c *AdminController) Routes(r kitrest.Router) {
	r.Method(http.MethodGet, "/api/admin/users", c.gate("users.read",
		kitrest.NewJsonApiListHandler(c.users, api.UserToDTO)))
	r.Method(http.MethodPost, "/api/admin/users", c.gate("users.write",
		kitrest.NewJsonApiCommandHandler(c.users.Create, decodeCreateUser, api.UserToDTO)))
	r.Method(http.MethodPatch, "/api/admin/users/{id}", c.gate("users.write",
		kitrest.NewJsonApiCommandHandler(c.users.Patch, decodePatchUser, api.UserToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodGet, "/api/admin/users/{id}/roles", c.gate("users.read",
		http.HandlerFunc(c.getUserRoles)))
	r.Method(http.MethodPut, "/api/admin/users/{id}/roles", c.gate("users.write",
		kitrest.NewJsonApiCommandHandler(c.users.SetRoles, decodeSetUserRoles, api.UserToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))

	r.Method(http.MethodGet, "/api/admin/roles", c.gate("roles.read",
		kitrest.NewJsonApiListHandler(c.roles, api.RoleToDTO)))
	r.Method(http.MethodGet, "/api/admin/roles/{id}", c.gate("roles.read",
		kitrest.NewJsonApiGetHandler(c.roles, api.RoleToDTO, []query.ParseOpt{})))
	r.Method(http.MethodPost, "/api/admin/roles", c.gate("roles.write",
		kitrest.NewJsonApiCommandHandler(c.roles.Upsert, decodeUpsertRole, api.RoleToDTO,
			kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	r.Method(http.MethodDelete, "/api/admin/roles/{id}", c.gate("roles.write",
		http.HandlerFunc(c.deleteRole)))

	r.Method(http.MethodGet, "/api/admin/permissions", c.gate("roles.read",
		kitrest.NewJsonApiListHandler(c.perms, api.PermissionToDTO)))

	r.Method(http.MethodGet, "/api/admin/apps", c.gate("apps.read",
		kitrest.NewJsonApiListHandler(c.apps, api.AppToDTO)))
	r.Method(http.MethodGet, "/api/admin/apps/{id}", c.gate("apps.read",
		kitrest.NewJsonApiGetHandler(c.apps, api.AppToDTO, []query.ParseOpt{})))
	for _, m := range []string{http.MethodPost, http.MethodPut} {
		r.Method(m, "/api/admin/apps/{id}", c.gate("apps.write",
			kitrest.NewJsonApiCommandHandler(c.apps.Upsert, decodeUpsertApp, api.AppToDTO,
				kitrest.HandlerWithSuccessStatus(http.StatusOK))))
	}
}

func decodeCreateUser(req *http.Request) (app.CreateUserCommand, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.UserCreateDTO](req)
	if err != nil {
		return app.CreateUserCommand{}, err
	}
	return app.CreateUserCommand{
		Email:       body.Email(),
		DisplayName: body.DisplayName(),
		Password:    body.Password(),
	}, nil
}

func decodePatchUser(req *http.Request) (app.PatchUserCommand, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.UserPatchDTO](req)
	if err != nil {
		return app.PatchUserCommand{}, err
	}
	return app.PatchUserCommand{
		UserID: req.PathValue("id"),
		Status: app.UserStatus(body.Status()),
	}, nil
}

func decodeSetUserRoles(req *http.Request) (app.SetUserRolesCommand, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.SetRolesDTO](req)
	if err != nil {
		return app.SetUserRolesCommand{}, err
	}
	return app.SetUserRolesCommand{
		UserID: req.PathValue("id"),
		Roles:  body.Roles(),
	}, nil
}

func decodeUpsertRole(req *http.Request) (app.UpsertRoleCommand, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.RoleUpsertDTO](req)
	if err != nil {
		return app.UpsertRoleCommand{}, err
	}
	return app.UpsertRoleCommand{
		Slug:        body.Slug(),
		Name:        body.Name(),
		Permissions: body.Permissions(),
	}, nil
}

func decodeUpsertApp(req *http.Request) (app.UpsertAppCommand, error) {
	body, err := kitrest.UnmarshalPayloadFromRequest[*api.AppUpsertDTO](req)
	if err != nil {
		return app.UpsertAppCommand{}, err
	}
	return app.UpsertAppCommand{
		Slug:         req.PathValue("id"),
		Name:         body.Name(),
		Kind:         body.Kind(),
		AdminBaseURL: body.AdminBaseURL(),
		Enabled:      body.Enabled(),
	}, nil
}

func (c *AdminController) getUserRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.users.RolesForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeRolesJSONAPI(w, roles)
}

func (c *AdminController) deleteRole(w http.ResponseWriter, r *http.Request) {
	if err := c.roles.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
