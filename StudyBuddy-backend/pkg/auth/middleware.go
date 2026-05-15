package auth

import (
	"net/http"
	"strings"
	"studybuddy/backend/pkg/httputil"
)

// Middleware validates Bearer JWT and sets user ID in request context.
// Use auth.UserIDFromContext(r.Context()) in handlers.
func Middleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				httputil.Error(w, http.StatusUnauthorized, "missing authorization")
				return
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				httputil.Error(w, http.StatusUnauthorized, "invalid authorization")
				return
			}
			token := strings.TrimSpace(auth[len(prefix):])
			claims, err := ValidateAccess(secret, token)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := WithUserID(r.Context(), claims.UserID)
			ctx = WithAccessClaims(ctx, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
