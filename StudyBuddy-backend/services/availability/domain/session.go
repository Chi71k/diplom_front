package domain

import (
	"errors"
	"time"
)

type Session struct {
	ID               string
	Title            string
	OrganizerID      string
	ParticipantIDs   []string
	CourseID         string
	GroupID          string
	StartTime        time.Time
	EndTime          time.Time
	Timezone         string
	Status           SessionStatus
	GCalEventIDs     map[string]string // userID → gcal event ID
	ParticipantsMeta []SessionParticipant
	CreatedAt        time.Time
}

// SessionParticipant captures membership and confirmation for API and use cases.
type SessionParticipant struct {
	UserID      string
	Confirmed   bool
	GCalEventID string
}

// SessionParticipantGCal is a participant row that has (or had) a linked Google Calendar event.
type SessionParticipantGCal struct {
	UserID      string
	GCalEventID string
}

type SessionStatus string

const (
	SessionProposed  SessionStatus = "proposed"
	SessionConfirmed SessionStatus = "confirmed"
	SessionCanceled  SessionStatus = "canceled"
)

var (
	ErrSessionInPast   = errors.New("session cannot start in the past")
	ErrInvalidDuration = errors.New("session duration must be between 15 minutes and 8 hours")
	ErrNoParticipants  = errors.New("session must have at least one other participant")
	ErrNotOrganizer    = errors.New("only the organizer can cancel this session")
	ErrSessionNotFound = errors.New("session not found")
	ErrNotParticipant  = errors.New("user is not a participant of this session")
)
