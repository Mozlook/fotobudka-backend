package router

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
)

func registerSessionRoutes(mux *http.ServeMux, sessionsHandler *sessions.Handler, manager *appauth.Manager) {
	mux.Handle("GET /api/sessions/{sessionId}", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.GetSessionByID)))
	mux.Handle("POST /api/sessions", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.InsertSession)))
	mux.Handle("GET /api/sessions", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.GetSessions)))
}
