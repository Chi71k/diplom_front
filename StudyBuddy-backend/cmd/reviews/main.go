package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"studybuddy/backend/pkg/db"
	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/reviews/delivery"
	"studybuddy/backend/services/reviews/repository"
	"studybuddy/backend/services/reviews/usecase"
)

func main() {
	_ = godotenv.Load(".env")

	port := getEnv("REVIEWS_SERVER_PORT", "8086")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	pointsURL := getEnv("POINTS_SERVICE_URL", "")
	dsn := getEnv("DATABASE_URL", "postgres://studybuddy:studybuddy@localhost:5432/studybuddy?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	reviewRepo := repository.NewPostgresReviewRepository(pool)
	matchChecker := repository.NewPostgresMatchChecker(pool)
	embCache := embedding.NewPgCache(pool)
	cacheInv := repository.NewEmbeddingCacheInvalidatorAdapter(embCache)

	createUC := usecase.NewCreateReview(reviewRepo, matchChecker, cacheInv, pointsURL, []byte(jwtSecret))
	listUC := usecase.NewListReviewsForUser(reviewRepo)
	ratingUC := usecase.NewGetAverageRating(reviewRepo)

	h := &delivery.ReviewsHandler{
		Create:    createUC,
		List:      listUC,
		GetRating: ratingUC,
	}
	router := delivery.NewRouter(h, []byte(jwtSecret))

	log.Printf("reviews service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
