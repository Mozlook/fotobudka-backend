package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// statusRecorder wraps http.ResponseWriter and records the final HTTP status code
// written by a handler.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// WriteHeader records the response status code and forwards it to the wrapped writer.
func (sr *statusRecorder) WriteHeader(code int) {
	if sr.wroteHeader {
		return
	}
	sr.status = code
	sr.wroteHeader = true
	sr.ResponseWriter.WriteHeader(code)
}

// Write ensures that an implicit HTTP 200 status is recorded when the handler
// writes a response body without calling WriteHeader first.
func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.wroteHeader == false {
		sr.WriteHeader(http.StatusOK)
	}
	return sr.ResponseWriter.Write(b)
}

// AccessLog logs one http_request event for every handled request.
func AccessLog(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		_, after, found := strings.Cut(r.Pattern, " ")

		path := r.Pattern
		if found {
			path = after
		}

		if path == "" {
			path = "/unmatched"
		}

		log.Info().
			Str("event_type", "http_request").
			Str("request_id", RequestIDFromContext(r.Context())).
			Str("http_method", r.Method).
			Str("http_path", path).
			Str("user_agent", r.UserAgent()).
			Str("src_ip", r.RemoteAddr).
			Int("http_status", rec.status).
			Float64("latency_ms", float64(time.Since(start).Microseconds())/1000).
			Msg("request")
	})
}
