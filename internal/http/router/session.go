package router

import (
	"net/http"

	sessions "github.com/Mozlook/fotobudka-backend/internal/http/handler/session"
)

func registerSessionRoutes(mux *http.ServeMux, sessionsHandler *sessions.Handler) {
	mux.Handle("GET /api/me/profile", http.HandlerFunc(sessionsHandler.GetSession))
}
