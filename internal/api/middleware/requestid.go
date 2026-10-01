package middleware

import (
	"context"
	"net/http"

	"github.com/madhavbiju/homelabd/internal/ctxutil"

	"github.com/google/uuid"
)





// RequestID adds a unique UUID to each request context and sets the X-Request-ID header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), ctxutil.RequestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

