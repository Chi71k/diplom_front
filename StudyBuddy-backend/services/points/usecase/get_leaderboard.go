package usecase

import (
	"context"

	"studybuddy/backend/services/points/domain"
)

type GetLeaderboard interface {
	GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error)
}

type getLeaderboard struct {
	repo PointsRepository
}

func NewGetLeaderboard(repo PointsRepository) GetLeaderboard {
	return &getLeaderboard{repo: repo}
}

func (uc *getLeaderboard) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	return uc.repo.ListLeaderboard(ctx, limit)
}
