package usecase

import (
	"context"

	"studybuddy/backend/services/reviews/domain"
)

type GetAverageRating interface {
	GetAverageRating(ctx context.Context, userID string) (domain.RatingSummary, error)
}

type getAverageRating struct {
	repo ReviewRepository
}

func NewGetAverageRating(repo ReviewRepository) GetAverageRating {
	return &getAverageRating{repo: repo}
}

func (uc *getAverageRating) GetAverageRating(ctx context.Context, userID string) (domain.RatingSummary, error) {
	return uc.repo.GetAverageRating(ctx, userID)
}
