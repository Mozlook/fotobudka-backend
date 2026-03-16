package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

const maxStacktraceLen = 8192

func Recover(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			stacktrace := string(debug.Stack())
			if len(stacktrace) > maxStacktraceLen {
				stacktrace = stacktrace[:maxStacktraceLen]
			}

			if rec := recover(); rec != nil {
				log.Error().
					Str("event_type", "unhandled_exception").
					Str("error_type", "UnhandledException").
					Str("request_id", RequestIDFromContext(r.Context())).
					Str("src_ip", r.RemoteAddr).
					Str("http_method", r.Method).
					Str("http_path", r.URL.Path).
					Dict("data", zerolog.Dict().
						Str("panic", fmt.Sprint(rec)).
						Str("stacktrace", stacktrace),
					).Msg("panic recovered")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
