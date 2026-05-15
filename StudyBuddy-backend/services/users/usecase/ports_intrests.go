package usecase

import "context"
import "studybuddy/backend/services/users/domain"

// InterestRepository for listing and resolving interests.
type InterestRepository interface {
	ListAll(ctx context.Context) ([]domain.Interest, error)
	GetByIDs(ctx context.Context, ids []string) ([]domain.Interest, error)
}

type UserInterestRepository interface {
	ListForUser(ctx context.Context, userID string) ([]domain.Interest, error)
	ReplaceForUser(ctx context.Context, userID string, interestIDs []string) error
}
