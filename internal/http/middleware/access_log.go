package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	if sr.wroteHeader {
		return
	}
	sr.status = code
	sr.wroteHeader = true
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.wroteHeader == false {
		sr.WriteHeader(http.StatusOK)
	}
	return sr.ResponseWriter.Write(b)
}

func AccessLog(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		log.Info().
			Str("event_type", "http_request").
			Str("request_id", RequestIDFromContext(r.Context())).
			Str("http_method", r.Method).
			Str("http_path", r.URL.Path).
			Str("user_agent", r.UserAgent()).
			Str("src_ip", r.RemoteAddr).
			Int("http_status", rec.status).
			Float64("latency_ms", float64(time.Since(start).Microseconds())/1000).
			Msg("request")
	})
}
