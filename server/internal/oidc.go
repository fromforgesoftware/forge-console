package internal

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/fromforgesoftware/go-kit/auth/oidc"
	"github.com/fromforgesoftware/go-kit/monitoring/logger"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// oidcProviderConfig is one entry of FOUNDRY_OIDC_PROVIDERS (a JSON array).
type oidcProviderConfig struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Issuer       string   `json:"issuer"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	Scopes       []string `json:"scopes"`
}

// newOIDCProviders builds the external-login providers from the environment.
// Unset/invalid config yields no providers — the console still works with the
// local password login.
func newOIDCProviders() app.OIDCProviders {
	raw := os.Getenv("FOUNDRY_OIDC_PROVIDERS")
	if raw == "" {
		return app.NewOIDCProviders(nil)
	}
	var cfgs []oidcProviderConfig
	if err := json.Unmarshal([]byte(raw), &cfgs); err != nil {
		logger.New().Error("invalid FOUNDRY_OIDC_PROVIDERS json", "error", err)
		return app.NewOIDCProviders(nil)
	}
	providers := make([]app.OIDCProvider, 0, len(cfgs))
	for _, c := range cfgs {
		if c.ID == "" || c.Issuer == "" || c.ClientID == "" {
			continue
		}
		name := c.Name
		if name == "" {
			name = c.ID
		}
		providers = append(providers, app.OIDCProvider{
			ID:     c.ID,
			Name:   name,
			Issuer: strings.TrimRight(c.Issuer, "/"),
			Client: oidc.NewClient(oidc.Provider{
				Issuer:       c.Issuer,
				ClientID:     c.ClientID,
				ClientSecret: c.ClientSecret,
				Scopes:       c.Scopes,
			}, nil),
		})
	}
	return app.NewOIDCProviders(providers)
}
