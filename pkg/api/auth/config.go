package auth

import "strings"

// DevUserIDHeader is the development-only local user bypass header.
const DevUserIDHeader = "X-GameHub-Dev-User-ID"

// Config contains OAuth and OIDC settings.
type Config struct {
	// Provider identifies the configured identity provider preset.
	Provider string `mapstructure:"provider" default:"generic_oidc"`

	// IssuerURL is the trusted OIDC issuer URL.
	IssuerURL string `mapstructure:"issuer_url"`

	// Audience is the API audience expected in access tokens.
	Audience string `mapstructure:"audience"`

	// ClientID is the public frontend client identifier.
	ClientID string `mapstructure:"client_id"`

	// Scopes is the frontend-safe scope list.
	Scopes string `mapstructure:"scopes" default:"openid profile email"`

	// DevelopmentBypass enables local user ID bypass in development only.
	DevelopmentBypass bool `mapstructure:"development_bypass" default:"false"`
}

// Public contains frontend-safe auth configuration.
type Public struct {
	// Provider identifies the configured identity provider preset.
	Provider string `json:"provider"`

	// IssuerURL is the OIDC issuer URL.
	IssuerURL string `json:"issuer_url"`

	// Audience is the API audience.
	Audience string `json:"audience"`

	// ClientID is the public frontend client identifier.
	ClientID string `json:"client_id"`

	// Scopes is the requested frontend scope list.
	Scopes []string `json:"scopes"`
}

// Public returns frontend-safe configuration.
func (config Config) Public() Public {
	return Public{
		Provider:  config.Provider,
		IssuerURL: config.IssuerURL,
		Audience:  config.Audience,
		ClientID:  config.ClientID,
		Scopes:    config.ScopeList(),
	}
}

// ScopeList returns configured scopes split on whitespace.
func (config Config) ScopeList() []string {
	return strings.Fields(config.Scopes)
}
