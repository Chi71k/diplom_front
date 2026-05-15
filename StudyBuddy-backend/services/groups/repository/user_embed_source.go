package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// userEmbedSource loads user fields needed for semantic embeddings (same data as matching service).
type userEmbedSource struct {
	pool *pgxpool.Pool
}

type profileData struct {
	FirstName string
	LastName  string
	Bio       string
}

func newUserEmbedSource(pool *pgxpool.Pool) *userEmbedSource {
	return &userEmbedSource{pool: pool}
}

func (s *userEmbedSource) GetProfile(ctx context.Context, userID string) (*profileData, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(bio, '')
FROM users
WHERE id = $1 AND is_active = true;
`
	var p profileData
	err := s.pool.QueryRow(ctx, q, userID).Scan(&p.FirstName, &p.LastName, &p.Bio)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *userEmbedSource) ListInterestNames(ctx context.Context, userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
SELECT i.name
FROM user_interests ui
JOIN interests i ON i.id = ui.interest_id
WHERE ui.user_id = $1
ORDER BY i.name;
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list interest names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

func (s *userEmbedSource) ListCourseTitles(ctx context.Context, userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
SELECT title
FROM courses
WHERE owner_user_id = $1
ORDER BY title;
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list course titles: %w", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, rows.Err()
}
