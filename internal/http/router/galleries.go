package router

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	gallerieshandler "github.com/Mozlook/fotobudka-backend/internal/http/handler/galleries"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
)

func RegisterGalleryRoutes(
	mux *http.ServeMux,
	authManager *appauth.Manager,
	galleriesHandler *gallerieshandler.Handler,
) {
	mux.Handle("GET /api/galleries", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.ListGalleries)))
	mux.Handle("POST /api/galleries", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.CreateGallery)))
	mux.Handle("GET /api/galleries/{galleryId}", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.GetGallery)))
	mux.Handle("PUT /api/galleries/{galleryId}", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.UpdateGallery)))
	mux.Handle("DELETE /api/galleries/{galleryId}", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.DeleteGallery)))
	mux.Handle("POST /api/galleries/{galleryId}/photos/presign", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.PresignGalleryPhotos)))
	mux.Handle("POST /api/galleries/{galleryId}/photos/{photoId}/complete", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.CompleteGalleryPhoto)))
	mux.Handle("DELETE /api/galleries/{galleryId}/photos/{photoId}", middleware.RequireAuth(authManager, http.HandlerFunc(galleriesHandler.DeleteGalleryPhoto)))
}
