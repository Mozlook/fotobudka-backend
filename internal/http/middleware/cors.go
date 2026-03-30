package middleware

import (
	"net/http"
	"strings"
)

// CORS adds cross-origin headers for the configured frontend origin
// and handles preflight OPTIONS requests.
func CORS(frontendOrigin string, next http.Handler) http.Handler {
	allowedOrigin := strings.TrimSpace(frontendOrigin)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		addVary(w.Header(), "Origin")
		addVary(w.Header(), "Access-Control-Request-Method")
		addVary(w.Header(), "Access-Control-Request-Headers")

		if origin == "" || origin != allowedOrigin {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func addVary(h http.Header, value string) {
	existing := h.Values("Vary")
	for _, v := range existing {
		for _, part := range strings.Split(v, ",") {
			if strings.TrimSpace(part) == value {
				return
			}
		}
	}
	h.Add("Vary", value)
}
