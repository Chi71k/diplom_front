package usecase

import (
	"context"

	"studybuddy/backend/services/users/domain"
)

// AdminStats returns platform-wide aggregate statistics.
type AdminStats interface {
	Execute(ctx context.Context) (domain.PlatformStats, error)
}

type adminStats struct {
	repo AdminRepository
}

// NewAdminStats creates the AdminStats use case.
func NewAdminStats(repo AdminRepository) AdminStats {
	return &adminStats{repo: repo}
}

// Execute loads dashboard stats from the repository.
func (uc *adminStats) Execute(ctx context.Context) (domain.PlatformStats, error) {
	return uc.repo.GetPlatformStats(ctx)
}
