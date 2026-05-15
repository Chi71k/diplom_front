package usecase

import "context"
import "studybuddy/backend/services/availability/domain"

type DeleteSlotInput struct {
	UserID string // from JWT, used for ownership check
	SlotID string
}

type DeleteSlot interface {
	DeleteSlot(ctx context.Context, input DeleteSlotInput) error
}

type deleteSlot struct {
	repo SlotRepository
}

func NewDeleteSlot(repo SlotRepository) DeleteSlot {
	return &deleteSlot{repo: repo}
}

func (d *deleteSlot) DeleteSlot(ctx context.Context, input DeleteSlotInput) error {
	existing, err := d.repo.GetByID(ctx, input.SlotID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrSlotNotFound
	}
	// Ownership check: only the owner can delete their slot
	if existing.UserID != input.UserID {
		return domain.ErrForbidden
	}
	return d.repo.Delete(ctx, input.SlotID)
}
