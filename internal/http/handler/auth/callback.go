package auth

import "net/http"

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	rawQuery := r.URL.Query()
	code := rawQuery.Get("code")
	if code == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	state := rawQuery.Get("state")
	if state == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	stateCookie, err := r.Cookie("fotobudka_oauth_state")
	if err != nil || state != stateCookie.Value {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	verifierCookie, err := r.Cookie("fotobudka_oauth_verifier")
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	_, err = h.provider.Exchange(r.Context(), code, verifierCookie.Value)
	if err != nil {
		h.clearFlowCookies(w, stateCookie.Name)
		h.clearFlowCookies(w, verifierCookie.Name)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	h.clearFlowCookies(w, stateCookie.Name)
	h.clearFlowCookies(w, verifierCookie.Name)
	w.WriteHeader(http.StatusOK)
	return
}
