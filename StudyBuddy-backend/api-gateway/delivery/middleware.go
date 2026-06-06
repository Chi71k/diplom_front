package delivery

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	pkgauth "studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
)

type ctxKey string

const requestIDKey ctxKey = "requestID"

// RequestIDFromContext returns the correlation ID set by RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// Recovery recovers panics, logs them, and returns 500 JSON.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic request_id=%s path=%s err=%v", RequestIDFromContext(r.Context()), r.URL.Path, rec)
				httputil.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestID assigns X-Request-ID (UUID) when missing and stores it on the context/response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestLogger logs method, path, status, duration, and request-id after the response completes.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		log.Printf(
			"request method=%s path=%s status=%d duration_ms=%d request_id=%s",
			r.Method,
			r.URL.Path,
			ww.Status(),
			time.Since(start).Milliseconds(),
			RequestIDFromContext(r.Context()),
		)
	})
}

// CORS returns chi/cors middleware configured from allowed origins.
// github.com/go-chi/cors — no CORS exists in-repo today.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// RateLimiterConfig holds per-IP token bucket limits.
type RateLimiterConfig struct {
	GlobalRPM float64
	AuthRPM   float64
}

// RateLimiter enforces in-memory per-IP limits (global and stricter auth paths).
// golang.org/x/time/rate — stdlib token bucket; no external rate-limit service.
type RateLimiter struct {
	globalRPM rate.Limit
	authRPM   rate.Limit
	mu        sync.Mutex
	global    map[string]*rate.Limiter
	auth      map[string]*rate.Limiter
}

// NewRateLimiter builds limiters from requests-per-minute settings.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	global := cfg.GlobalRPM
	if global <= 0 {
		global = 100
	}
	auth := cfg.AuthRPM
	if auth <= 0 {
		auth = 20
	}
	return &RateLimiter{
		globalRPM: rate.Limit(global / 60.0),
		authRPM:   rate.Limit(auth / 60.0),
		global:    make(map[string]*rate.Limiter),
		auth:      make(map[string]*rate.Limiter),
	}
}

func (rl *RateLimiter) getLimiter(m map[string]*rate.Limiter, key string, limit rate.Limit) *rate.Limiter {
	lim, ok := m[key]
	if !ok {
		burst := int(limit * 60)
		if burst < 1 {
			burst = 1
		}
		if burst > 200 {
			burst = 200
		}
		lim = rate.NewLimiter(limit, burst)
		m[key] = lim
	}
	return lim
}

// Middleware returns HTTP middleware that returns 429 when the client exceeds its bucket.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		rl.mu.Lock()
		lim := rl.getLimiter(rl.global, ip, rl.globalRPM)
		if isAuthRateLimitPath(r.URL.Path) {
			lim = rl.getLimiter(rl.auth, ip, rl.authRPM)
		}
		rl.mu.Unlock()
		if !lim.Allow() {
			httputil.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAuthRateLimitPath(path string) bool {
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/register":
		return true
	default:
		return false
	}
}

// clientIP returns the client address, preferring the first X-Forwarded-For hop.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// AuthGuard validates JWT on protected /api/v1 routes and forwards Authorization unchanged.
func AuthGuard(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if !requiresAuth(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			if auth == "" {
				httputil.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				httputil.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			token := strings.TrimSpace(auth[len(prefix):])
			if _, err := pkgauth.ValidateAccess(jwtSecret, token); err != nil {
				httputil.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// requiresAuth is true for protected /api/v1 routes per gateway auth rules.
func requiresAuth(path string) bool {
	if !strings.HasPrefix(path, "/api/v1/") {
		return false
	}
	return !isPublicAPIPath(path)
}

// isPublicAPIPath reports routes that skip JWT validation at the gateway.
func isPublicAPIPath(path string) bool {
	switch path {
	case "/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/availability/gcal/callback":
		return true
	}
	if isPublicReviewsRating(path) || isPublicUserPoints(path) {
		return true
	}
	return false
}

func isPublicReviewsRating(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 6 {
		return false
	}
	return parts[0] == "api" && parts[1] == "v1" && parts[2] == "reviews" &&
		parts[3] == "users" && parts[5] == "rating"
}

func isPublicUserPoints(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 5 {
		return false
	}
	return parts[0] == "api" && parts[1] == "v1" && parts[2] == "users" && parts[4] == "points"
}
