package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"studybuddy/backend/api-gateway/delivery"
)

func main() {
	_ = godotenv.Load(".env")

	port := getEnv("GATEWAY_PORT", "8000")
	jwtSecret := []byte(getEnv("JWT_SECRET", "dev-secret-change-in-production"))
	if len(jwtSecret) < 32 {
		log.Print("warning: JWT_SECRET should be at least 32 bytes for production")
	}

	shutdownTimeout := getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	healthTimeout := getEnvDuration("HEALTH_CHECK_TIMEOUT", 3*time.Second)

	cfg := delivery.GatewayConfig{
		JWTSecret: jwtSecret,
		ServiceURLs: delivery.ServiceURLs{
			Auth:         getEnv("AUTH_SERVICE_URL", "http://auth:8080"),
			Users:        getEnv("USERS_SERVICE_URL", "http://users:8081"),
			Courses:      getEnv("COURSES_SERVICE_URL", "http://courses:8082"),
			Availability: getEnv("AVAILABILITY_SERVICE_URL", "http://availability:8083"),
			Matching:     getEnv("MATCHING_SERVICE_URL", "http://matching:8084"),
			Groups:       getEnv("GROUPS_SERVICE_URL", "http://groups:8085"),
			Reviews:      getEnv("REVIEWS_SERVICE_URL", "http://reviews:8086"),
			Points:       getEnv("POINTS_SERVICE_URL", "http://points:8087"),
		},
		CORSOrigins:      splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173")),
		RateLimitRPM:     getEnvFloat("RATE_LIMIT_RPM", 100),
		RateLimitAuthRPM: getEnvFloat("RATE_LIMIT_AUTH_RPM", 20),
		HealthTimeout:    healthTimeout,
	}

	router, err := delivery.NewRouter(cfg)
	if err != nil {
		log.Fatalf("router: %v", err)
	}

	addr := ":" + port
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("api-gateway listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Print("api-gateway shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
		os.Exit(1)
	}
	log.Print("api-gateway stopped")
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && f > 0 {
			return f
		}
	}
	return defaultVal
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
