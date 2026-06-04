package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// BackfillUser is a user row eligible for embedding backfill.
type BackfillUser struct {
	ID        string
	FirstName string
	LastName  string
	Bio       string
}

// ListUsersNeedingEmbeddingBackfill returns active users without a stored embedding vector.
func (r *PgProfileRepository) ListUsersNeedingEmbeddingBackfill(ctx context.Context) ([]BackfillUser, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	const q = `
SELECT id::text, first_name, last_name, COALESCE(bio, '')
FROM users
WHERE embedding_vector IS NULL AND is_active = true
ORDER BY created_at ASC;
`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BackfillUser
	for rows.Next() {
		var u BackfillUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Bio); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// UpdateEmbedding persists a 768-dimensional embedding vector for
// the given user. Uses pgvector's native vector type via SQL cast.
func (r *PgProfileRepository) UpdateEmbedding(ctx context.Context, userID string, vector []float32) error {
	if len(vector) != 768 {
		return fmt.Errorf("embedding must have 768 dimensions, got %d", len(vector))
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `UPDATE users SET embedding_vector = $1::vector WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, formatVector(vector), userID)
	return err
}

// GetEmbedding retrieves the stored embedding vector for a user.
// Returns nil slice (not an error) if no embedding has been generated yet.
func (r *PgProfileRepository) GetEmbedding(ctx context.Context, userID string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT embedding_vector::text FROM users WHERE id = $1`
	var raw *string
	err := r.pool.QueryRow(ctx, q, userID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if raw == nil || *raw == "" {
		return nil, nil
	}
	return parseVectorText(*raw)
}

// ListInterestNamesForUser returns interest names for embedding narrative.
func (r *PgProfileRepository) ListInterestNamesForUser(ctx context.Context, userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT i.name
FROM user_interests ui
JOIN interests i ON i.id = ui.interest_id
WHERE ui.user_id = $1
ORDER BY i.name;
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
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

// ListCourseTitlesForUser returns course titles for embedding narrative.
func (r *PgProfileRepository) ListCourseTitlesForUser(ctx context.Context, userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT title
FROM courses
WHERE owner_user_id = $1
ORDER BY title;
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
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

func formatVector(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

func parseVectorText(s string) ([]float32, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		f, err := strconv.ParseFloat(p, 32)
		if err != nil {
			return nil, fmt.Errorf("parse vector component %q: %w", p, err)
		}
		out = append(out, float32(f))
	}
	return out, nil
}
