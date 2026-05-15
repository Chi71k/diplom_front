package usecase

import (
	"context"

	"studybuddy/backend/services/availability/domain"
)

type ListMySessions interface {
	ListMySessions(ctx context.Context, userID string) ([]domain.Session, error)
}

type listMySessions struct {
	repo SessionRepository
}

func NewListMySessions(repo SessionRepository) ListMySessions {
	return &listMySessions{repo: repo}
}

func (uc *listMySessions) ListMySessions(ctx context.Context, userID string) ([]domain.Session, error) {
	return uc.repo.ListForUser(ctx, userID)
}
