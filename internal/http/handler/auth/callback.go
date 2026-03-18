package auth

import (
	"encoding/json"
	"net/http"
)

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	provider := h.provider

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

	token, err := provider.Exchange(r.Context(), code, verifierCookie.Value)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	userData, err := provider.FetchUserInfo(r.Context(), token)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	payload, err := json.Marshal(userData)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.clearFlowCookies(w, stateCookie.Name)
	h.clearFlowCookies(w, verifierCookie.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
