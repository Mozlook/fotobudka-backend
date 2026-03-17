package handler

import (
	"crypto/rand"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/Mozlook/fotobudka-backend/internal/oauth"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	cfg      config.Config
	provider *oauth.Provider
}

func NewAuthHandler(cfg config.Config, provider *oauth.Provider) *AuthHandler {
	return &AuthHandler{
		cfg:      cfg,
		provider: provider,
	}
}

func (h *AuthHandler) setFlowCookies(w http.ResponseWriter, name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   600,
	}

	http.SetCookie(w, cookie)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	provider := h.provider
	verifier := oauth2.GenerateVerifier()
	state := rand.Text()

	h.setFlowCookies(w, "fotobudka_oauth_verifier", verifier)
	h.setFlowCookies(w, "fotobudka_oauth_state", state)

	loginURL := provider.LoginURL(state, verifier)
	http.Redirect(w, r, loginURL, http.StatusFound)
}
