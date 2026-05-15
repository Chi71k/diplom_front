package usecase

import (
	"context"

	"studybuddy/backend/services/points/domain"
)

type MyPointsResult struct {
	TotalPoints  int64
	Transactions []domain.Transaction
}

type GetMyPoints interface {
	GetMyPoints(ctx context.Context, userID string) (*MyPointsResult, error)
}

type getMyPoints struct {
	repo PointsRepository
}

func NewGetMyPoints(repo PointsRepository) GetMyPoints {
	return &getMyPoints{repo: repo}
}

func (uc *getMyPoints) GetMyPoints(ctx context.Context, userID string) (*MyPointsResult, error) {
	total, err := uc.repo.SumForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	txs, err := uc.repo.ListRecentForUser(ctx, userID, 20)
	if err != nil {
		return nil, err
	}
	return &MyPointsResult{TotalPoints: total, Transactions: txs}, nil
}
