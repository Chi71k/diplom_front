package gcal

import (
	"context"

	"google.golang.org/api/calendar/v3"

	"studybuddy/backend/services/availability/domain"
)

// CalendarService returns a Google Calendar API client for the connection.
func (p *Provider) CalendarService(ctx context.Context, conn *domain.GCalConnection) (*calendar.Service, error) {
	return p.calendarService(ctx, conn)
}
