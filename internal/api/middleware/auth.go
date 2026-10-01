package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/madhavbiju/homelabd/internal/api/response"
	"github.com/madhavbiju/homelabd/internal/ctxutil"
	"github.com/madhavbiju/homelabd/internal/auth"
)





// RequireAuth ensures the request has a valid Bearer token.
func RequireAuth(authService *auth.Service, requiredPerms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Missing or invalid Authorization header")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			record, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				response.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
				return
			}

			// Check permissions if any are required
			if len(requiredPerms) > 0 {
				hasPerm := false
				for _, rp := range requiredPerms {
					for _, up := range record.Permissions {
						if up == "admin" || up == rp { // 'admin' scope overrides all
							hasPerm = true
							break
						}
					}
					if hasPerm {
						break
					}
				}

				if !hasPerm {
					response.WriteError(w, r, http.StatusForbidden, "forbidden", "Insufficient permissions")
					return
				}
			}

			ctx := context.WithValue(r.Context(), ctxutil.TokenRecordKey, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

