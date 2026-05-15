package usecase

import (
	"context"

	"studybuddy/backend/services/reviews/domain"
)

type ListReviewsForUser interface {
	ListReviewsForUser(ctx context.Context, userID string) ([]domain.Review, error)
}

type listReviewsForUser struct {
	repo ReviewRepository
}

func NewListReviewsForUser(repo ReviewRepository) ListReviewsForUser {
	return &listReviewsForUser{repo: repo}
}

func (uc *listReviewsForUser) ListReviewsForUser(ctx context.Context, userID string) ([]domain.Review, error) {
	return uc.repo.ListForReviewee(ctx, userID)
}
