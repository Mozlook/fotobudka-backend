package auth

import (
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/repository/users"
	"github.com/google/uuid"
)

// GoogleCallback completes the Google OAuth login flow.
//
// It validates the returned state, exchanges the authorization code for
// an OAuth token, fetches the Google user profile, upserts the local user,
// issues the application auth token, clears temporary flow cookies,
// and redirects the client to the frontend application.
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	stateCookie, err := r.Cookie("fotobudka_oauth_state")
	if err != nil || stateCookie.Value == "" || state != stateCookie.Value {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	verifierCookie, err := r.Cookie("fotobudka_oauth_verifier")
	if err != nil || verifierCookie.Value == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := h.provider.Exchange(r.Context(), code, verifierCookie.Value)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	userData, err := h.provider.FetchUserInfo(r.Context(), token)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	upsert := users.UpsertFromGoogleInput{
		ID:        uuid.New(),
		GoogleSub: userData.Sub,
		Email:     userData.Email,
		Name:      userData.Name,
		AvatarURL: userData.Picture,
	}

	user, err := h.users.UpsertFromGoogle(r.Context(), upsert)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tokenStr, expiresAt, err := h.manager.IssueToken(user.ID.String())
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.manager.SetAuthCookie(w, tokenStr, expiresAt)
	h.clearFlowCookies(w, stateCookie.Name)
	h.clearFlowCookies(w, verifierCookie.Name)

	http.Redirect(w, r, h.cfg.HTTP.FrontendOrigin, http.StatusFound)
}
