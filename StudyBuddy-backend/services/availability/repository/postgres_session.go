package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/availability/domain"
	"studybuddy/backend/services/availability/usecase"
)

type PgSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPgSessionRepository(pool *pgxpool.Pool) usecase.SessionRepository {
	return &PgSessionRepository{pool: pool}
}

func (r *PgSessionRepository) Create(ctx context.Context, title, organizerID string, participantUserIDs []string, courseID, groupID string, start, end time.Time, timezone string) (*domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var courseArg, groupArg any
	if courseID != "" {
		courseArg = courseID
	}
	if groupID != "" {
		groupArg = groupID
	}
	if timezone == "" {
		timezone = "UTC"
	}

	const insS = `
INSERT INTO study_sessions (title, organizer_id, course_id, group_id, start_time, end_time, timezone, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'proposed')
RETURNING id::text, created_at;
`
	var s domain.Session
	err = tx.QueryRow(ctx, insS, title, organizerID, courseArg, groupArg, start, end, timezone).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	s.Title = title
	s.OrganizerID = organizerID
	s.CourseID = courseID
	s.GroupID = groupID
	s.StartTime = start
	s.EndTime = end
	s.Timezone = timezone
	s.Status = domain.SessionProposed

	const insP = `
INSERT INTO session_participants (session_id, user_id, confirmed)
VALUES ($1, $2, $3);
`
	if _, err := tx.Exec(ctx, insP, s.ID, organizerID, true); err != nil {
		return nil, err
	}
	for _, uid := range participantUserIDs {
		if uid == "" || uid == organizerID {
			continue
		}
		if _, err := tx.Exec(ctx, insP, s.ID, uid, false); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, s.ID)
}

func (r *PgSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	const q = `
SELECT id::text, title, organizer_id::text,
       COALESCE(course_id::text, ''), COALESCE(group_id::text, ''),
       start_time, end_time, timezone, status, created_at
FROM study_sessions
WHERE id = $1;
`
	var s domain.Session
	var status string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.Title, &s.OrganizerID, &s.CourseID, &s.GroupID,
		&s.StartTime, &s.EndTime, &s.Timezone, &status, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Status = domain.SessionStatus(status)

	rows, err := r.pool.Query(ctx, `
SELECT user_id::text, confirmed, COALESCE(gcal_event_id, '')
FROM session_participants
WHERE session_id = $1
ORDER BY CASE WHEN user_id = $2::uuid THEN 0 ELSE 1 END, user_id;
`, id, s.OrganizerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s.GCalEventIDs = make(map[string]string)
	for rows.Next() {
		var p domain.SessionParticipant
		if err := rows.Scan(&p.UserID, &p.Confirmed, &p.GCalEventID); err != nil {
			return nil, err
		}
		s.ParticipantsMeta = append(s.ParticipantsMeta, p)
		s.ParticipantIDs = append(s.ParticipantIDs, p.UserID)
		if p.GCalEventID != "" {
			s.GCalEventIDs[p.UserID] = p.GCalEventID
		}
	}
	return &s, rows.Err()
}

func (r *PgSessionRepository) ListForUser(ctx context.Context, userID string) ([]domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
SELECT s.id::text
FROM study_sessions s
JOIN session_participants sp ON sp.session_id = s.id AND sp.user_id = $1
ORDER BY s.start_time ASC;
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []domain.Session
	for _, id := range ids {
		s, err := r.GetByID(ctx, id)
		if err != nil || s == nil {
			continue
		}
		out = append(out, *s)
	}
	return out, nil
}

func (r *PgSessionRepository) SetParticipantConfirmed(ctx context.Context, sessionID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.pool.Exec(ctx, `
UPDATE session_participants SET confirmed = true WHERE session_id = $1 AND user_id = $2;
`, sessionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotParticipant
	}
	return nil
}

func (r *PgSessionRepository) AllParticipantsConfirmed(ctx context.Context, sessionID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var ok bool
	err := r.pool.QueryRow(ctx, `
SELECT NOT EXISTS (
  SELECT 1 FROM session_participants WHERE session_id = $1 AND confirmed = false
);
`, sessionID).Scan(&ok)
	return ok, err
}

func (r *PgSessionRepository) SetSessionStatus(ctx context.Context, sessionID string, status domain.SessionStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.pool.Exec(ctx, `
UPDATE study_sessions SET status = $2 WHERE id = $1;
`, sessionID, string(status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *PgSessionRepository) UpdateParticipantGCalEvent(ctx context.Context, sessionID, userID, eventID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx, `
UPDATE session_participants SET gcal_event_id = $3 WHERE session_id = $1 AND user_id = $2;
`, sessionID, userID, eventID)
	return err
}

func (r *PgSessionRepository) ListParticipantsWithGCalEvents(ctx context.Context, sessionID string) ([]domain.SessionParticipantGCal, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
SELECT user_id::text, gcal_event_id
FROM session_participants
WHERE session_id = $1 AND gcal_event_id IS NOT NULL AND gcal_event_id <> '';
`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list gcal participants: %w", err)
	}
	defer rows.Close()

	var out []domain.SessionParticipantGCal
	for rows.Next() {
		var row domain.SessionParticipantGCal
		if err := rows.Scan(&row.UserID, &row.GCalEventID); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
