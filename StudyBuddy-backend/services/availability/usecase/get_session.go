package usecase

import (
	"context"

	"studybuddy/backend/services/availability/domain"
)

type GetSessionInput struct {
	SessionID string
}

type GetSession interface {
	GetSession(ctx context.Context, in GetSessionInput) (*domain.Session, error)
}

type getSession struct {
	repo SessionRepository
}

func NewGetSession(repo SessionRepository) GetSession {
	return &getSession{repo: repo}
}

func (uc *getSession) GetSession(ctx context.Context, in GetSessionInput) (*domain.Session, error) {
	s, err := uc.repo.GetByID(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}
