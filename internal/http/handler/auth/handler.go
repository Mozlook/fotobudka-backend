package auth

import (
	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/Mozlook/fotobudka-backend/internal/oauth"
	"github.com/Mozlook/fotobudka-backend/internal/repository/users"
)

// AuthHandler serves HTTP endpoints related to authentication.
type AuthHandler struct {
	cfg      config.Config
	provider *oauth.Provider
	users    *users.Repository
	manager  *appauth.Manager
}

// NewAuthHandler creates an AuthHandler configured with application settings
// and the Google OAuth provider.
func NewAuthHandler(cfg config.Config, provider *oauth.Provider, repo *users.Repository, manager *appauth.Manager) *AuthHandler {
	return &AuthHandler{
		cfg:      cfg,
		provider: provider,
		users:    repo,
		manager:  manager,
	}
}
