package router

import (
	"net/http"

	publicportfoliohandler "github.com/Mozlook/fotobudka-backend/internal/http/handler/publicportfolio"
)

func RegisterPublicPortfolioRoutes(
	mux *http.ServeMux,
	publicPortfolioHandler *publicportfoliohandler.Handler,
) {
	mux.HandleFunc("GET /api/public/photographers/{username}", publicPortfolioHandler.GetPhotographer)
	mux.HandleFunc("GET /api/public/photographers/{username}/galleries/{slug}", publicPortfolioHandler.GetGallery)
}
