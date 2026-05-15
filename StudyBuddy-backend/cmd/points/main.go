package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"studybuddy/backend/pkg/db"
	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/points/delivery"
	"studybuddy/backend/services/points/repository"
	"studybuddy/backend/services/points/usecase"
)

func main() {
	_ = godotenv.Load(".env")

	port := getEnv("POINTS_SERVER_PORT", "8087")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	geminiKey := getEnv("GEMINI_API_KEY", "")
	reviewsURL := strings.TrimSpace(os.Getenv("REVIEWS_SERVICE_URL"))
	dsn := getEnv("DATABASE_URL", "postgres://studybuddy:studybuddy@localhost:5432/studybuddy?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	pointsRepo := repository.NewPostgresPointsRepository(pool)
	embCache := embedding.NewPgCache(pool)
	gemini := embedding.NewClient(geminiKey)
	embedProv := repository.NewEmbeddingProvider(gemini, embCache, pool)

	var reviewReader usecase.ReviewRatingReader = repository.NoopReviewsRating{}
	if reviewsURL != "" {
		reviewReader = repository.NewReviewsRatingClient(reviewsURL)
	}

	addUC := usecase.NewAddTransaction(pointsRepo)
	getMineUC := usecase.NewGetMyPoints(pointsRepo)
	getBoardUC := usecase.NewGetLeaderboard(pointsRepo)
	getRepUC := usecase.NewGetReputationLeaderboard(pointsRepo, reviewReader)
	searchUC := usecase.NewSearchLeaderboard(pointsRepo, gemini, embedProv, getBoardUC)
	publicUC := usecase.NewGetUserPointsTotal(pointsRepo)

	h := &delivery.PointsHandler{
		AddTx:          addUC,
		GetMine:        getMineUC,
		GetBoard:       getBoardUC,
		GetRepBoard:    getRepUC,
		SearchBoard:    searchUC,
		GetPublicTotal: publicUC,
	}
	router := delivery.NewRouter(h, []byte(jwtSecret))

	log.Printf("points service listening on :%s", port)
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
