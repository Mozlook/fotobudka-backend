package auth

import (
	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/Mozlook/fotobudka-backend/internal/oauth"
)

// AuthHandler serves HTTP endpoints related to authentication.
type AuthHandler struct {
	cfg      config.Config
	provider *oauth.Provider
}

// NewAuthHandler creates an AuthHandler configured with application settings
// and the Google OAuth provider.
func NewAuthHandler(cfg config.Config, provider *oauth.Provider) *AuthHandler {
	return &AuthHandler{
		cfg:      cfg,
		provider: provider,
	}
}
