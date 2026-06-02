package app

import "github.com/fromforgesoftware/go-kit/auth/oidc"

// OIDCProvider is one configured external login option (e.g. Google, GitHub, or
// an Aegis realm). Client is the kit OIDC relying-party for it.
type OIDCProvider struct {
	ID     string
	Name   string
	Issuer string // realm issuer base, for RP-initiated logout (issuer + /logout)
	Client *oidc.Client
}

// OIDCProviders is the immutable set of configured providers, preserving config
// order for stable login-button rendering.
type OIDCProviders struct {
	byID  map[string]OIDCProvider
	order []string
}

func NewOIDCProviders(list []OIDCProvider) OIDCProviders {
	byID := make(map[string]OIDCProvider, len(list))
	order := make([]string, 0, len(list))
	for _, p := range list {
		byID[p.ID] = p
		order = append(order, p.ID)
	}
	return OIDCProviders{byID: byID, order: order}
}

func (p OIDCProviders) Get(id string) (OIDCProvider, bool) {
	v, ok := p.byID[id]
	return v, ok
}

// List returns the providers in config order (for the login screen).
func (p OIDCProviders) List() []OIDCProvider {
	out := make([]OIDCProvider, 0, len(p.order))
	for _, id := range p.order {
		out = append(out, p.byID[id])
	}
	return out
}
