package auth

import "context"

type contextKey string

const (
	userContextKey   contextKey = "user"
	claimsContextKey contextKey = "accessClaims"
)

// WithUserID stores user ID in context (for use after JWT validation).
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userContextKey, userID)
}

// UserIDFromContext returns the user ID from context, or "" if missing.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userContextKey).(string)
	return v
}

// WithAccessClaims stores validated access JWT claims (set by auth.Middleware).
func WithAccessClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// AccessClaimsFromContext returns claims set by Middleware, or nil.
func AccessClaimsFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsContextKey).(*Claims)
	return v
}
