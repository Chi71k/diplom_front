package usecase

import "context"
import "studybuddy/backend/services/users/domain"

// ProfileRepository is the port for profile persistence.
type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	Upsert(ctx context.Context, profile *domain.Profile) error
	// DeleteByUserID performs logical deletion for the user (e.g. deactivate).
	DeleteByUserID(ctx context.Context, userID string) error
}
