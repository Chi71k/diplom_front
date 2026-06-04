package auth

import (
	"net/http"

	"studybuddy/backend/pkg/httputil"
)

// RequireAdmin wraps a handler and allows only JWTs with role "admin".
// It must run after Middleware, which validates the token and sets access claims.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := AccessClaimsFromContext(r.Context())
		if claims == nil || claims.Role != RoleAdmin {
			httputil.Error(w, http.StatusForbidden, "forbidden: admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
