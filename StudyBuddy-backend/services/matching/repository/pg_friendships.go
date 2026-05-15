package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/matching/domain"
	"studybuddy/backend/services/matching/usecase"
)

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// PgFriendshipRepository persists friendships (bidirectional rows).
type PgFriendshipRepository struct {
	pool *pgxpool.Pool
}

func NewPgFriendshipRepository(pool *pgxpool.Pool) usecase.FriendshipRepository {
	return &PgFriendshipRepository{pool: pool}
}

func execFriendshipsBothWays(ctx context.Context, e execer, userA, userB string) error {
	const q = `
INSERT INTO friendships (user_id, friend_id) VALUES ($1, $2), ($3, $4)
ON CONFLICT DO NOTHING;
`
	_, err := e.Exec(ctx, q, userA, userB, userB, userA)
	return err
}

func (r *PgFriendshipRepository) CreateBoth(ctx context.Context, userA, userB string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return execFriendshipsBothWays(ctx, r.pool, userA, userB)
}

func (r *PgFriendshipRepository) ListFriends(ctx context.Context, userID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT friend_id::text
FROM friendships
WHERE user_id = $1
ORDER BY created_at DESC;
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *PgFriendshipRepository) Delete(ctx context.Context, userID, friendID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
DELETE FROM friendships
WHERE (user_id = $1 AND friend_id = $2)
   OR (user_id = $2 AND friend_id = $1);
`
	tag, err := r.pool.Exec(ctx, q, userID, friendID)
	if err != nil {
		return fmt.Errorf("delete friendship: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFriendshipNotFound
	}
	return nil
}

func (r *PgFriendshipRepository) MutualFriendCount(ctx context.Context, userA, userB string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT COUNT(*)::int
FROM friendships f1
JOIN friendships f2 ON f1.friend_id = f2.friend_id
WHERE f1.user_id = $1 AND f2.user_id = $2;
`
	var n int
	err := r.pool.QueryRow(ctx, q, userA, userB).Scan(&n)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("mutual friend count: %w", err)
	}
	return n, nil
}
