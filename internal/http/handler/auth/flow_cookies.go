package auth

import "net/http"

// setFlowCookies writes a short-lived HttpOnly cookie used during the OAuth flow.
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

// clearFlowCookies removes a short-lived OAuth flow cookie from the client
func (h *AuthHandler) clearFlowCookies(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		HttpOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
