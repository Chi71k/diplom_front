package usecase

import (
	"context"
	"time"

	"studybuddy/backend/services/availability/domain"
)

type SlotRepository interface {
	Create(ctx context.Context, slot *domain.Slot) error
	ListForUser(ctx context.Context, userID string) ([]domain.Slot, error)
	GetByID(ctx context.Context, id string) (*domain.Slot, error)
	Delete(ctx context.Context, id string) error
	// DeleteAllForUser removes every slot belonging to a user
	// Example: user disconnects GCal and opts out to clear imported slots,
	// account deletion
	DeleteAllForUser(ctx context.Context, userID string) error

	// ReplaceForUser atomically replaces all slots for a user with the provided slice
	// It is used by the GCal import flow to keep local slots in sync
	ReplaceForUser(ctx context.Context, userID string, slots []domain.Slot) error

	// TODO: implement this method for Matching service later
	ListForUsers(ctx context.Context, userIDs []string) ([]domain.Slot, error)
}

type GCalRepository interface {
	GetConnection(ctx context.Context, userID string) (*domain.GCalConnection, error)
	UpsertConnection(ctx context.Context, connection *domain.GCalConnection) error
	DeleteConnection(ctx context.Context, userID string) error
}

// GCalProvider is the port for Google Calendar API
type GCalProvider interface {
	// ExchangeCode exchanges OAuth code for tokens
	ExchangeCode(ctx context.Context, code string) (*domain.GCalConnection, error)
	// RefreshToken refreshes an expired access token
	RefreshToken(ctx context.Context, conn *domain.GCalConnection) (*domain.GCalConnection, error)
	// ImportEvents fetches busy slots from GCal and converts to []domain.Slot
	ImportEvents(ctx context.Context, conn *domain.GCalConnection, userID string) ([]domain.Slot, error)
	// GetAuthURL returns the OAuth redirect the URL for the frontend
	GetAuthURL(state string) string
	// UpsertSessionEvent creates or updates a Google Calendar event for a study session.
	// Returns the Google Calendar event ID.
	UpsertSessionEvent(ctx context.Context, conn *domain.GCalConnection, session *domain.Session, userID string) (string, error)
	// DeleteSessionEvent deletes a previously created event by its Google Calendar event ID.
	DeleteSessionEvent(ctx context.Context, conn *domain.GCalConnection, eventID string) error
}

// SessionRepository persists study sessions and participants.
type SessionRepository interface {
	Create(ctx context.Context, title, organizerID string, participantUserIDs []string, courseID, groupID string, start, end time.Time, timezone string) (*domain.Session, error)
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Session, error)
	SetParticipantConfirmed(ctx context.Context, sessionID, userID string) error
	AllParticipantsConfirmed(ctx context.Context, sessionID string) (bool, error)
	SetSessionStatus(ctx context.Context, sessionID string, status domain.SessionStatus) error
	UpdateParticipantGCalEvent(ctx context.Context, sessionID, userID, eventID string) error
	ListParticipantsWithGCalEvents(ctx context.Context, sessionID string) ([]domain.SessionParticipantGCal, error)
}
