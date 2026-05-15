package usecase

import (
	"context"
	"errors"
	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/services/auth/domain"
)

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	AccessToken string
	ExpiresAt   int64
}

type Refresh interface {
	Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error)
}

type refresh struct {
	repo UserRepository
	jwt  JWTIssuer
}

func NewRefresh(repo UserRepository, jwt JWTIssuer) Refresh {
	return &refresh{repo: repo, jwt: jwt}
}

func (u *refresh) Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error) {
	userID, _, err := u.jwt.ParseRefresh(in.RefreshToken)
	if err != nil {
		return RefreshOutput{}, auth.ErrInvalidToken
	}
	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return RefreshOutput{}, domain.ErrInvalidCreds
		}
		return RefreshOutput{}, err
	}
	if !user.IsActive {
		return RefreshOutput{}, domain.ErrUserInactive
	}
	access, _, expAt, err := u.jwt.IssuePair(user.ID, user.Email)
	if err != nil {
		return RefreshOutput{}, err
	}
	return RefreshOutput{AccessToken: access, ExpiresAt: expAt}, nil
}
