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
	mux.Handle("GET /api/sessions", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.GetAllSessions)))
	mux.Handle("DELETE /api/sessions/{sessionId}", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.CloseSession)))
	mux.Handle("POST /api/sessions/{sessionId}/access/regenerate", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.RegenerateSessionAccess)))
	mux.Handle("POST /api/sessions/{sessionId}/photos/presign", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.PhotosPresign)))
	mux.Handle("POST /api/sessions/{sessionId}/photos/{photoId}/complete", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.PhotosComplete)))
	mux.Handle("POST /api/sessions/{sessionId}/payment/mark-paid", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.MarkPaid)))
	mux.Handle("POST /api.sessions/{sessionId}/payment/finals/presign", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.PresignFinals)))
	mux.Handle("POST /api/sessions/{sessionId}/finals/{finalId}/complete", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.CompleteFinalPhotoUpload)))
	mux.Handle("POST /api/sessions/{sessionId}/deliveries/generate-zip", middleware.RequireAuth(manager, http.HandlerFunc(sessionsHandler.GenerateZIP)))
}
