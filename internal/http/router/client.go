package router

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/http/handler/client"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/selections"
)

func registerClientRouter(
	mux *http.ServeMux,
	clientHandler *client.Handler,
	clientManager *appauth.ClientManager,
	sessionsRepo *sessions.Repository,
	selections *selections.Service,
) {
	clientAccess := func(next http.Handler) http.Handler {
		return middleware.RequireClientSessionAccess(clientManager, sessionsRepo, next)
	}

	mux.Handle("GET /api/client/access/by-token/{token}", http.HandlerFunc(clientHandler.GetSessionByToken))
	mux.Handle("POST /api/client/access/by-code", http.HandlerFunc(clientHandler.GetSessionByCode))

	mux.Handle("GET /api/client/session/{sessionId}/photos",
		clientAccess(http.HandlerFunc(clientHandler.GetSessionPhotos)),
	)
	mux.Handle("GET /api/client/photos/{photoId}/proof-url",
		clientAccess(http.HandlerFunc(clientHandler.GetClientPhotoProofURL)),
	)
	mux.Handle("PUT /api/client/session/{sessionId}/selections", clientAccess(http.HandlerFunc(clientHandler.UpdateSelections)))
	mux.Handle("POST /api/client/session/{sessionId}/submit", clientAccess(http.HandlerFunc(clientHandler.SubmitSelection)))
}
