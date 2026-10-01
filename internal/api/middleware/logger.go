package middleware

import (
	"log/slog"
	"net/http"

	"github.com/madhavbiju/homelabd/internal/ctxutil"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs the start and end of each HTTP request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := ctxutil.GetRequestID(r.Context())

		rw := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"request_id", reqID,
			"client_ip", r.RemoteAddr,
		)
	})
}
