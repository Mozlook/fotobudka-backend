package auth

import "net/http"

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.manager.ClearAuthCookie(w)

	w.WriteHeader(http.StatusNoContent)
}
