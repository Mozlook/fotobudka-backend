package oauth

import (
	"github.com/Mozlook/fotobudka-backend/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Provider wraps the OAuth 2.0 configuration used for Google authentication.
type Provider struct {
	config oauth2.Config
}

// New creates a Google OAuth provider configured from application settings.
func New(cfg config.Config) *Provider {
	return &Provider{
		oauth2.Config{
			ClientID:     cfg.OAuth.GoogleClientID,
			ClientSecret: cfg.OAuth.GoogleClientSecret,
			RedirectURL:  cfg.OAuth.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

// LoginURL returns the Google authorization URL for the current login flow.
//
// The returned URL includes the provided state value and PKCE challenge.
func (p *Provider) LoginURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}
