package usecase

import "context"
import "studybuddy/backend/services/availability/domain"

type ListSlots interface {
	ListSlots(ctx context.Context, userID string) ([]domain.Slot, error)
}

type listSlots struct {
	repo SlotRepository
}

func NewListSlots(repo SlotRepository) ListSlots {
	return &listSlots{repo: repo}
}

func (l *listSlots) ListSlots(ctx context.Context, userID string) ([]domain.Slot, error) {
	return l.repo.ListForUser(ctx, userID)
}
