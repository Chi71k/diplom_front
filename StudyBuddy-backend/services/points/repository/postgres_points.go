package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/points/domain"
)

type PostgresPointsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPointsRepository(pool *pgxpool.Pool) *PostgresPointsRepository {
	return &PostgresPointsRepository{pool: pool}
}

func (r *PostgresPointsRepository) AddTransaction(ctx context.Context, userID string, amount int, reason domain.Reason, sourceKey string) error {
	const q = `
INSERT INTO point_transactions (user_id, amount, reason, source_key)
VALUES ($1::uuid, $2::int, $3::text, $4::text)
ON CONFLICT (user_id, reason, source_key) WHERE (source_key <> '') DO NOTHING
`
	_, err := r.pool.Exec(ctx, q, userID, amount, string(reason), sourceKey)
	if err != nil {
		return err
	}
	tryRefreshLeaderboardMV(r.pool)
	return nil
}

func (r *PostgresPointsRepository) SumForUser(ctx context.Context, userID string) (int64, error) {
	const q = `SELECT COALESCE(SUM(amount), 0)::bigint FROM point_transactions WHERE user_id = $1::uuid`
	var total int64
	err := r.pool.QueryRow(ctx, q, userID).Scan(&total)
	return total, err
}

func (r *PostgresPointsRepository) ListRecentForUser(ctx context.Context, userID string, limit int) ([]domain.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	const q = `
SELECT id::text, user_id::text, amount, reason, source_key, created_at
FROM point_transactions
WHERE user_id = $1::uuid
ORDER BY created_at DESC
LIMIT $2
`
	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Reason, &t.SourceKey, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *PostgresPointsRepository) ListLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 500 {
		limit = 500
	}
	const q = `
SELECT user_id, total_points, rank FROM (
  SELECT user_id::text, total_points::bigint,
         ROW_NUMBER() OVER (ORDER BY total_points DESC)::int AS rank
  FROM leaderboard_points
) x
ORDER BY total_points DESC
LIMIT $1
`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.LeaderboardEntry
	for rows.Next() {
		var row domain.LeaderboardEntry
		if err := rows.Scan(&row.UserID, &row.TotalPoints, &row.Rank); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *PostgresPointsRepository) ListAllLeaderboardTotals(ctx context.Context) ([]domain.LeaderboardEntry, error) {
	const q = `
SELECT user_id::text, total_points::bigint,
       ROW_NUMBER() OVER (ORDER BY total_points DESC)::int AS rank
FROM leaderboard_points
ORDER BY total_points DESC
`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.LeaderboardEntry
	for rows.Next() {
		var row domain.LeaderboardEntry
		if err := rows.Scan(&row.UserID, &row.TotalPoints, &row.Rank); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *PostgresPointsRepository) MaxLeaderboardPoints(ctx context.Context) (int64, error) {
	const q = `SELECT COALESCE(MAX(total_points), 0)::bigint FROM leaderboard_points`
	var m int64
	err := r.pool.QueryRow(ctx, q).Scan(&m)
	return m, err
}

func (r *PostgresPointsRepository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.pool.Ping(ctx)
}
