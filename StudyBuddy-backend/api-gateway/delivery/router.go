package delivery

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// GatewayConfig holds upstream base URLs and runtime settings for the gateway router.
type GatewayConfig struct {
	JWTSecret        []byte
	ServiceURLs      ServiceURLs
	CORSOrigins      []string
	RateLimitRPM     float64
	RateLimitAuthRPM float64
	HealthTimeout    time.Duration // filled by caller; zero uses health checker default
}

// ServiceURLs maps logical service names to base URLs (scheme + host + port).
type ServiceURLs struct {
	Auth         string
	Users        string
	Courses      string
	Availability string
	Matching     string
	Groups       string
	Reviews      string
	Points       string
}

// NewRouter builds the gateway Chi router with middleware and reverse proxies.
func NewRouter(cfg GatewayConfig) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(Recovery)
	r.Use(RequestID)
	r.Use(RequestLogger)
	r.Use(CORS(cfg.CORSOrigins))

	rl := NewRateLimiter(RateLimiterConfig{
		GlobalRPM: cfg.RateLimitRPM,
		AuthRPM:   cfg.RateLimitAuthRPM,
	})
	r.Use(rl.Middleware)
	r.Use(AuthGuard(cfg.JWTSecret))

	health := NewHealthChecker(map[string]string{
		"auth":         cfg.ServiceURLs.Auth,
		"users":        cfg.ServiceURLs.Users,
		"courses":      cfg.ServiceURLs.Courses,
		"availability": cfg.ServiceURLs.Availability,
		"matching":     cfg.ServiceURLs.Matching,
		"groups":       cfg.ServiceURLs.Groups,
		"reviews":      cfg.ServiceURLs.Reviews,
		"points":       cfg.ServiceURLs.Points,
	}, cfg.HealthTimeout)
	r.Get("/health", health.HandleHealth)

	proxyFor := func(base string) (http.Handler, error) {
		return NewReverseProxy(base)
	}
	mountProxy := func(pattern, base string) error {
		h, err := proxyFor(base)
		if err != nil {
			return fmt.Errorf("%s: %w", pattern, err)
		}
		r.Handle(pattern, h)
		return nil
	}
	// mountProxyExact registers both the collection path and /* (chi /* does not match the bare collection URL).
	mountProxyExact := func(exact, wildcard, base string) error {
		h, err := proxyFor(base)
		if err != nil {
			return fmt.Errorf("%s: %w", exact, err)
		}
		r.Handle(exact, h)
		r.Handle(wildcard, h)
		return nil
	}

	routes := []struct {
		pattern string
		base    string
	}{
		{"/api/v1/auth/*", cfg.ServiceURLs.Auth},
		{"/api/v1/users/{userID}/points", cfg.ServiceURLs.Points},
		{"/api/v1/users/*", cfg.ServiceURLs.Users},
		{"/api/v1/interests", cfg.ServiceURLs.Users},
		{"/api/v1/interests/*", cfg.ServiceURLs.Users},
		{"/api/v1/admin/*", cfg.ServiceURLs.Users},
		{"/api/v1/availability/*", cfg.ServiceURLs.Availability},
		{"/api/v1/matching/*", cfg.ServiceURLs.Matching},
		{"/api/v1/points/*", cfg.ServiceURLs.Points},
	}
	exactRoutes := []struct {
		exact    string
		wildcard string
		base     string
	}{
		{"/api/v1/courses", "/api/v1/courses/*", cfg.ServiceURLs.Courses},
		{"/api/v1/groups", "/api/v1/groups/*", cfg.ServiceURLs.Groups},
		{"/api/v1/sessions", "/api/v1/sessions/*", cfg.ServiceURLs.Availability},
		{"/api/v1/reviews", "/api/v1/reviews/*", cfg.ServiceURLs.Reviews},
	}

	for _, rt := range routes {
		if err := mountProxy(rt.pattern, rt.base); err != nil {
			return nil, err
		}
	}
	for _, rt := range exactRoutes {
		if err := mountProxyExact(rt.exact, rt.wildcard, rt.base); err != nil {
			return nil, err
		}
	}

	return r, nil
}
