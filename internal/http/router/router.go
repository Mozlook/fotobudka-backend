package router

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/me"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/rs/zerolog"
)

func New(
	log zerolog.Logger,
	authHandler *auth.AuthHandler,
	meHandler *me.Handler,
	manager *appauth.Manager,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handler.Health)

	registerAuthRoutes(mux, authHandler)
	registerMeRoutes(mux, meHandler, manager)
	registerSessionRoutes(mux, sessionsHandler)

	var h http.Handler = mux
	h = middleware.Recover(log, h)
	h = middleware.AccessLog(log, h)
	h = middleware.RequestID(h)

	return h
}
