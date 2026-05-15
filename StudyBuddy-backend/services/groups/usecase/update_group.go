package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

type UpdateGroupInput struct {
	ActorID     string
	GroupID     string
	Name        *string
	Description *string
	CourseIDs   *[]string
}

type UpdateGroup interface {
	UpdateGroup(ctx context.Context, in UpdateGroupInput) (*domain.Group, error)
}

type updateGroup struct {
	repo GroupRepository
}

func NewUpdateGroup(repo GroupRepository) UpdateGroup {
	return &updateGroup{repo: repo}
}

func (uc *updateGroup) UpdateGroup(ctx context.Context, in UpdateGroupInput) (*domain.Group, error) {
	g, err := uc.repo.GetByID(ctx, in.GroupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.ErrGroupNotFound
	}
	if g.OwnerID != in.ActorID {
		return nil, domain.ErrNotOwner
	}

	var namePtr, descPtr *string
	if in.Name != nil {
		n := *in.Name
		if len(n) == 0 || len(n) > 100 {
			return nil, domain.ErrInvalidGroupName
		}
		namePtr = &n
	}
	if in.Description != nil {
		d := *in.Description
		descPtr = &d
	}

	if namePtr == nil && descPtr == nil && in.CourseIDs == nil {
		return uc.repo.GetByID(ctx, in.GroupID)
	}

	if err := uc.repo.UpdateMetadataAndCourses(ctx, in.GroupID, namePtr, descPtr, in.CourseIDs); err != nil {
		return nil, err
	}
	out, err := uc.repo.GetByID(ctx, in.GroupID)
	if err != nil || out == nil {
		return nil, domain.ErrGroupNotFound
	}
	return out, nil
}
