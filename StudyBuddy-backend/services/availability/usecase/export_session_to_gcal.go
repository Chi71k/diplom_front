package usecase

import (
	"context"
	"fmt"

	"studybuddy/backend/services/availability/domain"
)

// ExportSessionToGCalOutput is returned after a successful calendar export.
type ExportSessionToGCalOutput struct {
	GCalEventID string
}

// ExportSessionToGCal manually exports a confirmed study session to the user's Google Calendar.
type ExportSessionToGCal interface {
	Execute(ctx context.Context, userID, sessionID string) (ExportSessionToGCalOutput, error)
}

type exportSessionToGCal struct {
	sessions SessionRepository
	gcal     GCalProvider
	gcalRepo GCalRepository
}

// NewExportSessionToGCal creates the ExportSessionToGCal use case.
func NewExportSessionToGCal(sessions SessionRepository, gcal GCalProvider, gcalRepo GCalRepository) ExportSessionToGCal {
	return &exportSessionToGCal{sessions: sessions, gcal: gcal, gcalRepo: gcalRepo}
}

// Execute exports the session to Google Calendar for the requesting participant.
func (uc *exportSessionToGCal) Execute(ctx context.Context, userID, sessionID string) (ExportSessionToGCalOutput, error) {
	s, err := uc.sessions.GetByID(ctx, sessionID)
	if err != nil {
		return ExportSessionToGCalOutput{}, err
	}
	if s == nil {
		return ExportSessionToGCalOutput{}, domain.ErrSessionNotFound
	}
	if s.Status != domain.SessionConfirmed {
		return ExportSessionToGCalOutput{}, domain.ErrSessionNotConfirmed
	}

	participantIdx := -1
	for i, p := range s.ParticipantsMeta {
		if p.UserID == userID {
			participantIdx = i
			break
		}
	}
	if participantIdx < 0 {
		return ExportSessionToGCalOutput{}, domain.ErrNotParticipant
	}

	if s.GCalEventIDs != nil {
		if existing := s.GCalEventIDs[userID]; existing != "" {
			return ExportSessionToGCalOutput{GCalEventID: existing}, nil
		}
	}
	if s.ParticipantsMeta[participantIdx].GCalEventID != "" {
		return ExportSessionToGCalOutput{GCalEventID: s.ParticipantsMeta[participantIdx].GCalEventID}, nil
	}

	conn, err := uc.gcalRepo.GetConnection(ctx, userID)
	if err != nil {
		return ExportSessionToGCalOutput{}, err
	}
	if conn == nil {
		return ExportSessionToGCalOutput{}, domain.ErrGCalNotConnected
	}
	if !conn.SyncEnabled {
		return ExportSessionToGCalOutput{}, domain.ErrGCalSyncDisabled
	}

	conn, err = EnsureFreshGCalConnection(ctx, uc.gcal, uc.gcalRepo, conn)
	if err != nil {
		return ExportSessionToGCalOutput{}, err
	}

	eventID, err := uc.gcal.UpsertSessionEvent(ctx, conn, s, userID)
	if err != nil || eventID == "" {
		return ExportSessionToGCalOutput{}, fmt.Errorf("upsert gcal session event: %w", err)
	}

	if err := uc.sessions.UpdateParticipantGCalEvent(ctx, sessionID, userID, eventID); err != nil {
		return ExportSessionToGCalOutput{}, err
	}

	return ExportSessionToGCalOutput{GCalEventID: eventID}, nil
}
