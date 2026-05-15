package usecase

import (
	"context"

	"studybuddy/backend/services/availability/domain"
)

type ConfirmSessionInput struct {
	UserID    string
	SessionID string
}

type ConfirmSession interface {
	ConfirmSession(ctx context.Context, in ConfirmSessionInput) (*domain.Session, error)
}

type confirmSession struct {
	sessions         SessionRepository
	gcal             GCalProvider
	gcalRepo         GCalRepository
	pointsServiceURL string
	jwtSecret        []byte
}

func NewConfirmSession(sessions SessionRepository, gcal GCalProvider, gcalRepo GCalRepository, pointsServiceURL string, jwtSecret []byte) ConfirmSession {
	return &confirmSession{sessions: sessions, gcal: gcal, gcalRepo: gcalRepo, pointsServiceURL: pointsServiceURL, jwtSecret: jwtSecret}
}

func (uc *confirmSession) participantIndex(s *domain.Session, userID string) int {
	for i, p := range s.ParticipantsMeta {
		if p.UserID == userID {
			return i
		}
	}
	return -1
}

func (uc *confirmSession) ConfirmSession(ctx context.Context, in ConfirmSessionInput) (*domain.Session, error) {
	s, err := uc.sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.ErrSessionNotFound
	}
	if uc.participantIndex(s, in.UserID) < 0 {
		return nil, domain.ErrNotParticipant
	}
	if s.Status == domain.SessionCanceled {
		return nil, domain.ErrSessionNotFound
	}
	for _, p := range s.ParticipantsMeta {
		if p.UserID == in.UserID && p.Confirmed {
			return s, nil
		}
	}

	if err := uc.sessions.SetParticipantConfirmed(ctx, in.SessionID, in.UserID); err != nil {
		return nil, err
	}

	all, err := uc.sessions.AllParticipantsConfirmed(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if !all {
		return uc.sessions.GetByID(ctx, in.SessionID)
	}
	fresh, err := uc.sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if fresh != nil && fresh.Status == domain.SessionConfirmed {
		return fresh, nil
	}

	if err := uc.sessions.SetSessionStatus(ctx, in.SessionID, domain.SessionConfirmed); err != nil {
		return nil, err
	}

	s, err = uc.sessions.GetByID(ctx, in.SessionID)
	if err != nil || s == nil {
		return nil, domain.ErrSessionNotFound
	}

	for _, p := range s.ParticipantsMeta {
		conn, err := uc.gcalRepo.GetConnection(ctx, p.UserID)
		if err != nil || conn == nil || !conn.SyncEnabled {
			continue
		}
		conn, err = EnsureFreshGCalConnection(ctx, uc.gcal, uc.gcalRepo, conn)
		if err != nil {
			continue
		}
		eventID, err := uc.gcal.UpsertSessionEvent(ctx, conn, s, p.UserID)
		if err != nil || eventID == "" {
			continue
		}
		_ = uc.sessions.UpdateParticipantGCalEvent(ctx, s.ID, p.UserID, eventID)
		if s.GCalEventIDs == nil {
			s.GCalEventIDs = make(map[string]string)
		}
		s.GCalEventIDs[p.UserID] = eventID
	}

	var allUserIDs []string
	for _, p := range s.ParticipantsMeta {
		allUserIDs = append(allUserIDs, p.UserID)
	}
	FireSessionConfirmedPoints(uc.pointsServiceURL, uc.jwtSecret, s.ID, allUserIDs)

	return uc.sessions.GetByID(ctx, in.SessionID)
}
