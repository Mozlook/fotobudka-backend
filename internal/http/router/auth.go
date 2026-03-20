package router

import (
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/http/handler/auth"
)

func registerAuthRoutes(mux *http.ServeMux, authHandler *auth.AuthHandler) {
	mux.HandleFunc("GET /api/auth/google/login", authHandler.GoogleLogin)
	mux.HandleFunc("GET /api/auth/google/callback", authHandler.GoogleCallback)
}
