package oauth

import (
	"github.com/Mozlook/fotobudka-backend/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Provider struct {
	Config oauth2.Config
}

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

func (p *Provider) LoginURL(state, verifier string) string {
	return p.Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}
