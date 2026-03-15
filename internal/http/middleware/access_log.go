package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func AccessLog(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
	})
}
