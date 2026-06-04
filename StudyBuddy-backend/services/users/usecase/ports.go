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

// EmbeddingRepository persists semantic embedding vectors on users.
type EmbeddingRepository interface {
	// UpdateEmbedding persists a 768-dimensional embedding vector for the given user.
	UpdateEmbedding(ctx context.Context, userID string, vector []float32) error
	// GetEmbedding returns nil when no vector has been stored yet.
	GetEmbedding(ctx context.Context, userID string) ([]float32, error)
}

// ProfileEmbeddingSource supplies profile fields for embedding narrative construction.
type ProfileEmbeddingSource interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	ListInterestNamesForUser(ctx context.Context, userID string) ([]string, error)
	ListCourseTitlesForUser(ctx context.Context, userID string) ([]string, error)
}

// AdminRepository supports admin-only user and platform operations.
type AdminRepository interface {
	ListUsers(ctx context.Context, filter domain.AdminUserFilter, limit, offset int) ([]domain.UserSummary, int, error)
	SetUserActive(ctx context.Context, userID string, active bool) error
	GetPlatformStats(ctx context.Context) (domain.PlatformStats, error)
}
