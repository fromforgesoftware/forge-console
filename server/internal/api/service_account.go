package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// ServiceAccountDTO is the jsonapi representation of a machine identity. The
// one-time client secret is never part of this DTO — it is returned only by the
// dedicated create endpoint.
type ServiceAccountDTO struct {
	resource.RestDTO

	RName       string `jsonapi:"attr,name"`
	RClientID   string `jsonapi:"attr,clientId"`
	RStatus     string `jsonapi:"attr,status"`
	RLastUsedAt string `jsonapi:"attr,lastUsedAt,omitempty"`
}

func ServiceAccountToDTO(sa app.ServiceAccount) *ServiceAccountDTO {
	if sa == nil {
		return nil
	}
	dto := &ServiceAccountDTO{
		RestDTO:   resource.ToRestDTO(sa),
		RName:     sa.Name(),
		RClientID: sa.ClientID(),
		RStatus:   string(sa.Status()),
	}
	if sa.LastUsedAt() != nil {
		dto.RLastUsedAt = sa.LastUsedAt().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	dto.RType = app.ResourceTypeServiceAccount
	return dto
}

// ServiceAccountCredentialsToDTO maps the one-time create result, including the
// plaintext secret shown to the caller exactly once.
func ServiceAccountCredentialsToDTO(creds app.ServiceAccountCredentials) *ServiceAccountCredentialsDTO {
	dto := &ServiceAccountCredentialsDTO{
		RestDTO:       resource.ToRestDTO(creds.ServiceAccount),
		RName:         creds.ServiceAccount.Name(),
		RClientID:     creds.ClientID,
		RClientSecret: creds.ClientSecret,
	}
	dto.RType = app.ResourceTypeServiceAccount
	return dto
}
