package gcal

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"

	pkggcal "studybuddy/backend/pkg/gcal"
	"studybuddy/backend/services/availability/domain"
)

// ExportSlotEvents creates recurring weekly Google Calendar events for each availability slot.
func ExportSlotEvents(provider *pkggcal.Provider, ctx context.Context, conn *domain.GCalConnection, slots []domain.Slot) (int, error) {
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	svc, err := provider.CalendarService(cctx, conn)
	if err != nil {
		return 0, err
	}

	calendarID := conn.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	count := 0
	for _, slot := range slots {
		loc, loadErr := time.LoadLocation(slot.Timezone)
		if loadErr != nil {
			loc = time.UTC
		}

		startDate := nextSlotWeekday(time.Now().In(loc), slot.DayOfWeek)
		startTime := time.Date(
			startDate.Year(), startDate.Month(), startDate.Day(),
			slot.StartTime.Hour(), slot.StartTime.Minute(), 0, 0, loc,
		)
		endTime := time.Date(
			startDate.Year(), startDate.Month(), startDate.Day(),
			slot.EndTime.Hour(), slot.EndTime.Minute(), 0, 0, loc,
		)

		event := &calendar.Event{
			Summary:     "Study Availability",
			Description: "Availability slot exported from StudyBuddy.",
			Start: &calendar.EventDateTime{
				DateTime: startTime.Format(time.RFC3339),
				TimeZone: slot.Timezone,
			},
			End: &calendar.EventDateTime{
				DateTime: endTime.Format(time.RFC3339),
				TimeZone: slot.Timezone,
			},
			Recurrence: []string{"RRULE:FREQ=WEEKLY"},
			Reminders: &calendar.EventReminders{
				UseDefault:      true,
				ForceSendFields: []string{"UseDefault"},
			},
		}

		if _, err := svc.Events.Insert(calendarID, event).Context(cctx).Do(); err != nil {
			return count, fmt.Errorf("gcal insert slot event: %w", err)
		}
		count++
	}

	return count, nil
}

// nextSlotWeekday returns today or the nearest future date matching the ISO weekday
// (0 = Monday, 6 = Sunday).
func nextSlotWeekday(from time.Time, isoDay int) time.Time {
	// time.Weekday: Sunday=0, Monday=1 … Saturday=6
	// ISO weekday:  Monday=0 … Sunday=6
	target := time.Weekday((isoDay + 1) % 7)
	for from.Weekday() != target {
		from = from.AddDate(0, 0, 1)
	}
	return from
}
