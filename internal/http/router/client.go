package router

import (
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/http/handler/client"
)

func registerClientRouter(mux *http.ServeMux, clientHandler *client.Handler) {
	mux.HandleFunc("GET /api/client/session/by-token/{token}", clientHandler.GetSessionByToken)
}
