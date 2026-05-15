package usecase

import (
	"context"
)

type GetUserPointsTotal interface {
	GetUserPointsTotal(ctx context.Context, userID string) (int64, error)
}

type getUserPointsTotal struct {
	repo PointsRepository
}

func NewGetUserPointsTotal(repo PointsRepository) GetUserPointsTotal {
	return &getUserPointsTotal{repo: repo}
}

func (uc *getUserPointsTotal) GetUserPointsTotal(ctx context.Context, userID string) (int64, error) {
	return uc.repo.SumForUser(ctx, userID)
}
