package usecase

import (
	"context"

	"studybuddy/backend/services/users/domain"
)

// GetUser returns a user's profile by ID (public lookup; no email in HTTP layer).
type GetUser interface {
	GetUser(ctx context.Context, userID string) (*domain.Profile, error)
}

type getUser struct {
	repo ProfileRepository
}

// NewGetUser creates the GetUser use case.
func NewGetUser(repo ProfileRepository) GetUser {
	return &getUser{repo: repo}
}

func (u *getUser) GetUser(ctx context.Context, userID string) (*domain.Profile, error) {
	return u.repo.GetByUserID(ctx, userID)
}
