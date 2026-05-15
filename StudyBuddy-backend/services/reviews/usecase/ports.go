package usecase

import (
	"context"

	"studybuddy/backend/services/reviews/domain"
)

type ReviewRepository interface {
	Create(ctx context.Context, reviewerID, revieweeID, matchID string, rating int, comment string) (*domain.Review, error)
	ListForReviewee(ctx context.Context, revieweeID string) ([]domain.Review, error)
	GetAverageRating(ctx context.Context, revieweeID string) (domain.RatingSummary, error)
}

type MatchChecker interface {
	HasAcceptedMatchBetween(ctx context.Context, matchID, reviewerID, revieweeID string) (bool, error)
}

type EmbeddingCacheInvalidator interface {
	DeleteUserCache(ctx context.Context, userID string) error
}

type PointsEventPoster interface {
	PostPointsEvent(ctx context.Context, userID, reason string, amount int, matchID string)
}
