package gcal

import (
	"context"

	pkggcal "studybuddy/backend/pkg/gcal"
	"studybuddy/backend/services/availability/domain"
)

// CalendarProvider extends the OAuth calendar provider with study-session event helpers.
type CalendarProvider struct {
	*pkggcal.Provider
}

// UpsertSessionEvent implements usecase.GCalProvider.
func (p *CalendarProvider) UpsertSessionEvent(ctx context.Context, conn *domain.GCalConnection, session *domain.Session, userID string) (string, error) {
	return UpsertSessionEvent(p.Provider, ctx, conn, session, userID)
}

// DeleteSessionEvent implements usecase.GCalProvider.
func (p *CalendarProvider) DeleteSessionEvent(ctx context.Context, conn *domain.GCalConnection, eventID string) error {
	return DeleteSessionEvent(p.Provider, ctx, conn, eventID)
}
