package usecase

import (
	"context"
	"fmt"
	"time"

	"studybuddy/backend/services/availability/domain"
)

type ProposeSessionInput struct {
	OrganizerID    string
	Title          string
	ParticipantIDs []string
	CourseID       string
	GroupID        string
	StartTime      time.Time
	EndTime        time.Time
	Timezone       string
}

type ProposeSession interface {
	ProposeSession(ctx context.Context, in ProposeSessionInput) (*domain.Session, error)
}

type proposeSession struct {
	sessions SessionRepository
	gcal     GCalProvider
	gcalRepo GCalRepository
}

func NewProposeSession(sessions SessionRepository, gcal GCalProvider, gcalRepo GCalRepository) ProposeSession {
	return &proposeSession{sessions: sessions, gcal: gcal, gcalRepo: gcalRepo}
}

func (uc *proposeSession) ProposeSession(ctx context.Context, in ProposeSessionInput) (*domain.Session, error) {
	if in.Title == "" {
		return nil, fmt.Errorf("title required")
	}
	seen := make(map[string]struct{})
	var invitees []string
	for _, id := range in.ParticipantIDs {
		if id == "" || id == in.OrganizerID {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		invitees = append(invitees, id)
	}
	if len(invitees) < 1 {
		return nil, domain.ErrNoParticipants
	}

	dur := in.EndTime.Sub(in.StartTime)
	if dur < 15*time.Minute || dur > 8*time.Hour {
		return nil, domain.ErrInvalidDuration
	}
	if !in.StartTime.After(time.Now()) {
		return nil, domain.ErrSessionInPast
	}

	tz := in.Timezone
	if tz == "" {
		tz = "UTC"
	}

	s, err := uc.sessions.Create(ctx, in.Title, in.OrganizerID, invitees, in.CourseID, in.GroupID, in.StartTime, in.EndTime, tz)
	if err != nil {
		return nil, err
	}

	conn, err := uc.gcalRepo.GetConnection(ctx, in.OrganizerID)
	if err != nil || conn == nil || !conn.SyncEnabled {
		return s, nil
	}

	conn, err = EnsureFreshGCalConnection(ctx, uc.gcal, uc.gcalRepo, conn)
	if err != nil {
		return s, nil
	}

	eventID, err := uc.gcal.UpsertSessionEvent(ctx, conn, s, in.OrganizerID)
	if err != nil || eventID == "" {
		return s, nil
	}
	_ = uc.sessions.UpdateParticipantGCalEvent(ctx, s.ID, in.OrganizerID, eventID)
	return uc.sessions.GetByID(ctx, s.ID)
}
