package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

type GetGroup interface {
	GetGroup(ctx context.Context, groupID string) (*domain.Group, error)
}

type getGroup struct {
	repo GroupRepository
}

func NewGetGroup(repo GroupRepository) GetGroup {
	return &getGroup{repo: repo}
}

func (uc *getGroup) GetGroup(ctx context.Context, groupID string) (*domain.Group, error) {
	g, err := uc.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.ErrGroupNotFound
	}
	return g, nil
}
