package usecase

import (
	"context"
	"fmt"
	"studybuddy/backend/services/availability/domain"
	"time"
)

type GCalImport interface {
	ImportFromGCal(ctx context.Context, userID string) ([]domain.Slot, error)
}

type gcalImport struct {
	gcal     GCalProvider
	gcalRepo GCalRepository
	slotRepo SlotRepository
}

func NewGCalImport(gcal GCalProvider, gcalRepo GCalRepository, slotRepo SlotRepository) GCalImport {
	return &gcalImport{
		gcal:     gcal,
		gcalRepo: gcalRepo,
		slotRepo: slotRepo,
	}
}

func (gc *gcalImport) ImportFromGCal(ctx context.Context, userID string) ([]domain.Slot, error) {
	conn, err := gc.gcalRepo.GetConnection(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get gcal connection: %w", err)
	}
	if conn == nil {
		return nil, domain.ErrGCalNotConnected
	}
	if !conn.SyncEnabled {
		return nil, domain.ErrGCalSyncDisabled
	}

	if conn.TokenExpiry.Before(time.Now().Add(60 * time.Second)) {
		conn, err = gc.gcal.RefreshToken(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("gcal token refresh: %w", domain.ErrGCalRefreshFailed)
		}
		if err := gc.gcalRepo.UpsertConnection(ctx, conn); err != nil {
			return nil, fmt.Errorf("persist refreshed token: %w", err)
		}
	}

	imported, err := gc.gcal.ImportEvents(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("import gcal events: %w", err)
	}

	for i := range imported {
		if err := imported[i].Validate(); err != nil {
			continue
		}
	}

	if err := gc.slotRepo.ReplaceForUser(ctx, userID, imported); err != nil {
		return nil, fmt.Errorf("replace slots: %w", err)
	}

	now := time.Now()
	conn.LastSyncedAt = &now
	if err := gc.gcalRepo.UpsertConnection(ctx, conn); err != nil {
		return imported, fmt.Errorf("update last_synced_at: %w", err)
	}

	return imported, nil
}
