package usecase

import (
	"context"

	"studybuddy/backend/services/points/domain"
)

type PointsRepository interface {
	AddTransaction(ctx context.Context, userID string, amount int, reason domain.Reason, sourceKey string) error
	SumForUser(ctx context.Context, userID string) (int64, error)
	ListRecentForUser(ctx context.Context, userID string, limit int) ([]domain.Transaction, error)
	ListLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error)
	ListAllLeaderboardTotals(ctx context.Context) ([]domain.LeaderboardEntry, error)
	MaxLeaderboardPoints(ctx context.Context) (int64, error)
}

// ReviewRatingReader fetches average rating for a user from the reviews service.
type ReviewRatingReader interface {
	AverageRating(ctx context.Context, userID string) (avg float64, totalReviews int, err error)
}

// EmbeddingProvider returns a user embedding vector or nil when unavailable.
type EmbeddingProvider interface {
	GetOrCompute(ctx context.Context, userID string) ([]float64, error)
}
