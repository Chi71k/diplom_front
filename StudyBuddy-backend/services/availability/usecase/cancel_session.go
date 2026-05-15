package usecase

import (
	"context"

	"studybuddy/backend/services/availability/domain"
)

type CancelSessionInput struct {
	ActorID   string
	SessionID string
}

type CancelSession interface {
	CancelSession(ctx context.Context, in CancelSessionInput) error
}

type cancelSession struct {
	sessions SessionRepository
	gcal     GCalProvider
	gcalRepo GCalRepository
}

func NewCancelSession(sessions SessionRepository, gcal GCalProvider, gcalRepo GCalRepository) CancelSession {
	return &cancelSession{sessions: sessions, gcal: gcal, gcalRepo: gcalRepo}
}

func (uc *cancelSession) CancelSession(ctx context.Context, in CancelSessionInput) error {
	s, err := uc.sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return err
	}
	if s == nil {
		return domain.ErrSessionNotFound
	}
	if s.OrganizerID != in.ActorID {
		return domain.ErrNotOrganizer
	}
	if s.Status == domain.SessionCanceled {
		return nil
	}

	if err := uc.sessions.SetSessionStatus(ctx, in.SessionID, domain.SessionCanceled); err != nil {
		return err
	}

	rows, err := uc.sessions.ListParticipantsWithGCalEvents(ctx, in.SessionID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		conn, err := uc.gcalRepo.GetConnection(ctx, row.UserID)
		if err != nil || conn == nil {
			continue
		}
		conn, err = EnsureFreshGCalConnection(ctx, uc.gcal, uc.gcalRepo, conn)
		if err != nil {
			continue
		}
		_ = uc.gcal.DeleteSessionEvent(ctx, conn, row.GCalEventID)
	}
	return nil
}
