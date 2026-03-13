package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

func Recover(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().
					Str("event_type", "unhandled_exception").
					Str("request_id", RequestIDFromContext(r.Context())).
					Interface("panic", rec).
					Str("stacktrace", string(debug.Stack())).
					Msg("panic recovered")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
