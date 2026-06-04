package usecase

import (
	"context"
	"errors"

	"studybuddy/backend/services/users/domain"
)

// AdminSetUserActive activates or deactivates a user account.
type AdminSetUserActive interface {
	Execute(ctx context.Context, userID string, active bool) error
}

type adminSetUserActive struct {
	repo AdminRepository
}

// NewAdminSetUserActive creates the AdminSetUserActive use case.
func NewAdminSetUserActive(repo AdminRepository) AdminSetUserActive {
	return &adminSetUserActive{repo: repo}
}

// Execute updates is_active for the given user.
func (uc *adminSetUserActive) Execute(ctx context.Context, userID string, active bool) error {
	err := uc.repo.SetUserActive(ctx, userID, active)
	if err != nil {
		return err
	}
	return nil
}

// IsUserNotFound reports whether err is domain.ErrUserNotFound.
func IsUserNotFound(err error) bool {
	return errors.Is(err, domain.ErrUserNotFound)
}
