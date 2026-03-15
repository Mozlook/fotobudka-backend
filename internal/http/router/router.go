package router

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/Mozlook/fotobudka-backend/internal/http/handler"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
)

func New(log zerolog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Health)

	var h http.Handler = mux
	h = middleware.Recover(log, h)
	h = middleware.AccessLog(log, h)
	h = middleware.RequestID(h)

	return h
}
