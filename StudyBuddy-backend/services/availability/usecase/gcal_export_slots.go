package usecase

import (
	"context"
	"fmt"

	"studybuddy/backend/services/availability/domain"
)

// GCalExportSlots exports the user's availability slots as recurring weekly events to Google Calendar.
type GCalExportSlots interface {
	ExportSlotsToGCal(ctx context.Context, userID string) (int, error)
}

type gcalExportSlots struct {
	gcal     GCalProvider
	gcalRepo GCalRepository
	slotRepo SlotRepository
}

// NewGCalExportSlots creates the GCalExportSlots use case.
func NewGCalExportSlots(gcal GCalProvider, gcalRepo GCalRepository, slotRepo SlotRepository) GCalExportSlots {
	return &gcalExportSlots{gcal: gcal, gcalRepo: gcalRepo, slotRepo: slotRepo}
}

func (uc *gcalExportSlots) ExportSlotsToGCal(ctx context.Context, userID string) (int, error) {
	conn, err := uc.gcalRepo.GetConnection(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get gcal connection: %w", err)
	}
	if conn == nil {
		return 0, domain.ErrGCalNotConnected
	}
	if !conn.SyncEnabled {
		return 0, domain.ErrGCalSyncDisabled
	}

	conn, err = EnsureFreshGCalConnection(ctx, uc.gcal, uc.gcalRepo, conn)
	if err != nil {
		return 0, err
	}

	slots, err := uc.slotRepo.ListForUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("list slots: %w", err)
	}

	if len(slots) == 0 {
		return 0, nil
	}

	count, err := uc.gcal.ExportSlots(ctx, conn, slots)
	if err != nil {
		return count, fmt.Errorf("export slots to gcal: %w", err)
	}

	return count, nil
}
