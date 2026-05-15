package gcal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"

	pkggcal "studybuddy/backend/pkg/gcal"
	"studybuddy/backend/services/availability/domain"
)

// UpsertSessionEvent creates or updates a Google Calendar event for a study session.
func UpsertSessionEvent(provider *pkggcal.Provider, ctx context.Context, conn *domain.GCalConnection, session *domain.Session, userID string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	svc, err := provider.CalendarService(cctx, conn)
	if err != nil {
		return "", err
	}

	calendarID := conn.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	desc := "StudyBuddy session."
	if session.CourseID != "" {
		desc = fmt.Sprintf("StudyBuddy session. Course: %s.", session.CourseID)
	}
	if session.GroupID != "" {
		desc = strings.TrimSpace(desc + fmt.Sprintf(" Group: %s.", session.GroupID))
	}

	loc, err := time.LoadLocation(session.Timezone)
	if err != nil {
		loc = time.UTC
	}
	startLocal := session.StartTime.In(loc)
	endLocal := session.EndTime.In(loc)

	event := &calendar.Event{
		Summary: session.Title,
		Start: &calendar.EventDateTime{
			DateTime: startLocal.Format(time.RFC3339),
			TimeZone: session.Timezone,
		},
		End: &calendar.EventDateTime{
			DateTime: endLocal.Format(time.RFC3339),
			TimeZone: session.Timezone,
		},
		Description: desc,
		Reminders: &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{Method: "popup", Minutes: 60},
				{Method: "popup", Minutes: 10},
				{Method: "email", Minutes: 1440},
			},
		},
	}

	existing := ""
	if session.GCalEventIDs != nil {
		existing = session.GCalEventIDs[userID]
	}

	if existing != "" {
		updated, err := svc.Events.Update(calendarID, existing, event).Context(cctx).Do()
		if err != nil {
			return "", fmt.Errorf("gcal events update: %w", err)
		}
		return updated.Id, nil
	}

	created, err := svc.Events.Insert(calendarID, event).Context(cctx).Do()
	if err != nil {
		return "", fmt.Errorf("gcal events insert: %w", err)
	}
	return created.Id, nil
}

// DeleteSessionEvent deletes a Google Calendar event by ID; 404/410 are treated as success.
func DeleteSessionEvent(provider *pkggcal.Provider, ctx context.Context, conn *domain.GCalConnection, eventID string) error {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	svc, err := provider.CalendarService(cctx, conn)
	if err != nil {
		return err
	}
	calendarID := conn.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}
	err = svc.Events.Delete(calendarID, eventID).Context(cctx).Do()
	if err == nil {
		return nil
	}
	var gerr *googleapi.Error
	if errors.As(err, &gerr) && (gerr.Code == 404 || gerr.Code == 410) {
		return nil
	}
	return fmt.Errorf("gcal events delete: %w", err)
}
