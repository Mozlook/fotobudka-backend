package router

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/client"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/me"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	sessionsrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/selections"
	"github.com/rs/zerolog"
)

func New(
	log zerolog.Logger,
	authHandler *auth.AuthHandler,
	meHandler *me.Handler,
	sessionsHandler *sessions.Handler,
	clientHandler *client.Handler,
	manager *appauth.Manager,
	clientManager *appauth.ClientManager,
	sessionsRepo *sessionsrepo.Repository,
	frontendOrigin string,
	selections *selections.Service,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handler.Health)

	registerAuthRoutes(mux, authHandler)
	registerMeRoutes(mux, meHandler, manager)
	registerSessionRoutes(mux, sessionsHandler, manager)
	registerClientRouter(mux, clientHandler, clientManager, sessionsRepo, selections)

	var h http.Handler = mux
	h = middleware.CORS(frontendOrigin, h)
	h = middleware.Recover(log, h)
	h = middleware.AccessLog(log, h)
	h = middleware.RequestID(h)

	return h
}
