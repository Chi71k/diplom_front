package usecase

import (
	"context"
	"fmt"
	"time"

	"studybuddy/backend/services/availability/domain"
)

// EnsureFreshGCalConnection refreshes the OAuth token when it expires within 60 seconds and persists it.
func EnsureFreshGCalConnection(ctx context.Context, gcal GCalProvider, repo GCalRepository, conn *domain.GCalConnection) (*domain.GCalConnection, error) {
	if conn == nil {
		return nil, nil
	}
	if conn.TokenExpiry.Before(time.Now().Add(60 * time.Second)) {
		refreshed, err := gcal.RefreshToken(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("gcal token refresh: %w", domain.ErrGCalRefreshFailed)
		}
		if err := repo.UpsertConnection(ctx, refreshed); err != nil {
			return nil, fmt.Errorf("persist refreshed token: %w", err)
		}
		return refreshed, nil
	}
	return conn, nil
}
