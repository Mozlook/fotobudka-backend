package router

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/Mozlook/fotobudka-backend/internal/http/handler"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
)

func New(log zerolog.Logger, authHandler *auth.AuthHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Health)
	mux.HandleFunc("GET /api/auth/google/login", authHandler.GoogleLogin)
	mux.HandleFunc("GET /api/auth/google/callback", authHandler.GoogleCallback)

	var h http.Handler = mux
	h = middleware.Recover(log, h)
	h = middleware.AccessLog(log, h)
	h = middleware.RequestID(h)

	return h
}
