package handler

import (
	"crypto/rand"
	"net/http"

	"golang.org/x/oauth2"
)

// GoogleLogin starts the Google OAuth login flow.
//
// It generates a fresh PKCE verifier and state value, stores them in short-lived
// flow cookies, and redirects the client to the Google authorization page.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	provider := h.provider
	verifier := oauth2.GenerateVerifier()
	state := rand.Text()

	h.setFlowCookies(w, "fotobudka_oauth_verifier", verifier)
	h.setFlowCookies(w, "fotobudka_oauth_state", state)

	loginURL := provider.LoginURL(state, verifier)
	http.Redirect(w, r, loginURL, http.StatusFound)
}
