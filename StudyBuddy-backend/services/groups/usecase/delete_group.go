package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

type DeleteGroupInput struct {
	ActorID string
	GroupID string
}

type DeleteGroup interface {
	DeleteGroup(ctx context.Context, in DeleteGroupInput) error
}

type deleteGroup struct {
	repo GroupRepository
}

func NewDeleteGroup(repo GroupRepository) DeleteGroup {
	return &deleteGroup{repo: repo}
}

func (uc *deleteGroup) DeleteGroup(ctx context.Context, in DeleteGroupInput) error {
	g, err := uc.repo.GetByID(ctx, in.GroupID)
	if err != nil {
		return err
	}
	if g == nil {
		return domain.ErrGroupNotFound
	}
	if g.OwnerID != in.ActorID {
		return domain.ErrNotOwner
	}
	return uc.repo.Delete(ctx, in.GroupID)
}
