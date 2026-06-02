package api

import "github.com/fromforgesoftware/go-kit/resource"

// UserCreateDTO is the JSON:API request body for POST /api/admin/users. The
// password is write-only and never echoed back.
type UserCreateDTO struct {
	resource.RestDTO

	REmail       string `jsonapi:"attr,email"`
	RDisplayName string `jsonapi:"attr,displayName,omitempty"`
	RPassword    string `jsonapi:"attr,password,omitempty"`
}

func (d *UserCreateDTO) Email() string       { return d.REmail }
func (d *UserCreateDTO) DisplayName() string { return d.RDisplayName }
func (d *UserCreateDTO) Password() string    { return d.RPassword }

// UserPatchDTO is the JSON:API request body for PATCH /api/admin/users/{id}.
type UserPatchDTO struct {
	resource.RestDTO

	RStatus string `jsonapi:"attr,status"`
}

func (d *UserPatchDTO) Status() string { return d.RStatus }

// RoleUpsertDTO is the JSON:API request body for POST /api/admin/roles.
type RoleUpsertDTO struct {
	resource.RestDTO

	RSlug        string   `jsonapi:"attr,slug"`
	RName        string   `jsonapi:"attr,name"`
	RPermissions []string `jsonapi:"attr,permissions"`
}

func (d *RoleUpsertDTO) Slug() string          { return d.RSlug }
func (d *RoleUpsertDTO) Name() string          { return d.RName }
func (d *RoleUpsertDTO) Permissions() []string { return d.RPermissions }

// AppUpsertDTO is the JSON:API request body for the admin registry editor.
type AppUpsertDTO struct {
	resource.RestDTO

	RName         string `jsonapi:"attr,name"`
	RKind         string `jsonapi:"attr,kind"`
	RAdminBaseURL string `jsonapi:"attr,adminBaseURL"`
	REnabled      bool   `jsonapi:"attr,enabled"`
}

func (d *AppUpsertDTO) Name() string         { return d.RName }
func (d *AppUpsertDTO) Kind() string         { return d.RKind }
func (d *AppUpsertDTO) AdminBaseURL() string { return d.RAdminBaseURL }
func (d *AppUpsertDTO) Enabled() bool        { return d.REnabled }

// ServiceAccountCreateDTO is the JSON:API request body for creating a service
// account.
type ServiceAccountCreateDTO struct {
	resource.RestDTO

	RName string `jsonapi:"attr,name"`
}

func (d *ServiceAccountCreateDTO) Name() string { return d.RName }

// ServiceAccountCredentialsDTO is returned once at creation with the one-time
// plaintext secret.
type ServiceAccountCredentialsDTO struct {
	resource.RestDTO

	RName         string `jsonapi:"attr,name"`
	RClientID     string `jsonapi:"attr,clientId"`
	RClientSecret string `jsonapi:"attr,clientSecret"`
}

// SetRolesDTO is the JSON:API request body for replacing a subject's roles.
type SetRolesDTO struct {
	resource.RestDTO

	RRoles []string `jsonapi:"attr,roles"`
}

func (d *SetRolesDTO) Roles() []string { return d.RRoles }

// MeProfilePatchDTO is the JSON:API body for PATCH /api/users/me (self-service
// profile edit) — a `users` resource carrying the editable display name.
type MeProfilePatchDTO struct {
	resource.RestDTO

	RDisplayName string `jsonapi:"attr,displayName"`
}

func (d *MeProfilePatchDTO) DisplayName() string { return d.RDisplayName }

// PasswordChangeDTO is the JSON:API body for PUT /api/users/me/password — a
// `users` resource carrying the write-only credential change.
type PasswordChangeDTO struct {
	resource.RestDTO

	RCurrentPassword string `jsonapi:"attr,currentPassword"`
	RNewPassword     string `jsonapi:"attr,newPassword"`
}

func (d *PasswordChangeDTO) CurrentPassword() string { return d.RCurrentPassword }
func (d *PasswordChangeDTO) NewPassword() string     { return d.RNewPassword }
