package usecase

import (
	"context"
	"strings"

	"studybuddy/backend/services/groups/domain"
)

type CreateGroupInput struct {
	OwnerID     string
	Name        string
	Description string
	CourseIDs   []string
}

type CreateGroup interface {
	CreateGroup(ctx context.Context, in CreateGroupInput) (*domain.Group, error)
}

type createGroup struct {
	repo GroupRepository
}

func NewCreateGroup(repo GroupRepository) CreateGroup {
	return &createGroup{repo: repo}
}

func (uc *createGroup) CreateGroup(ctx context.Context, in CreateGroupInput) (*domain.Group, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 100 {
		return nil, domain.ErrInvalidGroupName
	}
	desc := strings.TrimSpace(in.Description)
	return uc.repo.Create(ctx, name, desc, in.OwnerID, in.CourseIDs)
}
