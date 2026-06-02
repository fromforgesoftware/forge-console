package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// ResourceTypeSession is the login resource: a POST creates one (exchanging
// credentials for an httpOnly cookie); logout deletes it.
const ResourceTypeSession resource.Type = "sessions"

// ResourceTypeAuthProvider is an external OIDC provider the user may sign in
// through.
const ResourceTypeAuthProvider resource.Type = "auth-providers"

// SessionCreateDTO decodes a login request — a `sessions` resource carrying the
// credential attributes. Write-only; never marshaled back (the response is the
// authenticated `users` resource).
type SessionCreateDTO struct {
	resource.RestDTO

	REmail    string `jsonapi:"attr,email"`
	RPassword string `jsonapi:"attr,password"`
}

// AuthUserSettingsDTO is the nested settings attribute on the current-user view.
type AuthUserSettingsDTO struct {
	Theme string `json:"theme"`
}

// AuthUserDTO is the "current identity" view of a user — the SPA hydration
// payload: profile + effective roles/permissions + settings. Resource type is
// `users` (a fuller representation of the same resource the admin surface lists).
type AuthUserDTO struct {
	resource.RestDTO

	REmail       string              `jsonapi:"attr,email"`
	RDisplayName string              `jsonapi:"attr,displayName,omitempty"`
	RIsAdmin     bool                `jsonapi:"attr,isAdmin"`
	RRoles       []string            `jsonapi:"attr,roles"`
	RPermissions []string            `jsonapi:"attr,permissions"`
	RSettings    AuthUserSettingsDTO `jsonapi:"attr,settings"`
}

func AuthUserToDTO(u app.User, settings app.UserSettings, roles, permissions []string) *AuthUserDTO {
	isAdmin := false
	for _, p := range permissions {
		if p == "*.*" || p == "*" {
			isAdmin = true
			break
		}
	}
	dto := &AuthUserDTO{
		RestDTO:      resource.ToRestDTO(u),
		REmail:       u.Email(),
		RDisplayName: u.DisplayName(),
		RIsAdmin:     isAdmin,
		RRoles:       roles,
		RPermissions: permissions,
		RSettings:    AuthUserSettingsDTO{Theme: settings.Theme},
	}
	dto.RType = app.ResourceTypeUser
	return dto
}

// ResourceTypeUserSettings is the signed-in user's preferences singleton.
const ResourceTypeUserSettings resource.Type = "user-settings"

// UserSettingsDTO is the jsonapi representation of a user's preferences (read on
// GET, write on PUT /api/users/me/settings). Id is the user id.
type UserSettingsDTO struct {
	resource.RestDTO

	RTheme string `jsonapi:"attr,theme"`
}

func (d *UserSettingsDTO) Theme() string { return d.RTheme }

func UserSettingsToDTO(userID string, s app.UserSettings) *UserSettingsDTO {
	dto := &UserSettingsDTO{RTheme: s.Theme}
	dto.RID = userID
	dto.RType = ResourceTypeUserSettings
	return dto
}

// AuthProviderDTO is the jsonapi representation of a configured OIDC provider.
type AuthProviderDTO struct {
	resource.RestDTO

	RName string `jsonapi:"attr,name"`
}

func AuthProviderToDTO(id, name string) *AuthProviderDTO {
	dto := &AuthProviderDTO{RName: name}
	dto.RID = id
	dto.RType = ResourceTypeAuthProvider
	return dto
}
