package router

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/me"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
)

func registerMeRoutes(mux *http.ServeMux, meHandler *me.Handler, manager *appauth.Manager) {
	mux.Handle("GET /api/me/profile", middleware.RequireAuth(manager, http.HandlerFunc(meHandler.GetProfile)))
}
