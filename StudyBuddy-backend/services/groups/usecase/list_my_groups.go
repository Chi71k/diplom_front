package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

type ListMyGroups interface {
	ListMyGroups(ctx context.Context, userID string) ([]domain.Group, error)
}

type listMyGroups struct {
	repo GroupRepository
}

func NewListMyGroups(repo GroupRepository) ListMyGroups {
	return &listMyGroups{repo: repo}
}

func (uc *listMyGroups) ListMyGroups(ctx context.Context, userID string) ([]domain.Group, error) {
	return uc.repo.ListForUser(ctx, userID)
}
