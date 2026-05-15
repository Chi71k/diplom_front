package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

type RemoveMemberInput struct {
	ActorID  string
	GroupID  string
	TargetID string
}

type RemoveMember interface {
	RemoveMember(ctx context.Context, in RemoveMemberInput) error
}

type removeMember struct {
	repo GroupRepository
}

func NewRemoveMember(repo GroupRepository) RemoveMember {
	return &removeMember{repo: repo}
}

func (uc *removeMember) RemoveMember(ctx context.Context, in RemoveMemberInput) error {
	g, err := uc.repo.GetByID(ctx, in.GroupID)
	if err != nil {
		return err
	}
	if g == nil {
		return domain.ErrGroupNotFound
	}

	if in.TargetID == in.ActorID {
		if g.OwnerID == in.ActorID {
			return domain.ErrOwnerCannotLeave
		}
		return uc.repo.RemoveMember(ctx, in.GroupID, in.TargetID)
	}

	if g.OwnerID != in.ActorID {
		return domain.ErrNotOwner
	}
	return uc.repo.RemoveMember(ctx, in.GroupID, in.TargetID)
}
