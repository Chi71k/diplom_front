package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMatchChecker struct {
	pool *pgxpool.Pool
}

func NewPostgresMatchChecker(pool *pgxpool.Pool) *PostgresMatchChecker {
	return &PostgresMatchChecker{pool: pool}
}

func (r *PostgresMatchChecker) HasAcceptedMatchBetween(ctx context.Context, matchID, reviewerID, revieweeID string) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1 FROM matches
  WHERE id = $1::uuid
    AND status = 'accepted'
    AND (
      (requester_id = $2::uuid AND receiver_id = $3::uuid)
      OR (requester_id = $3::uuid AND receiver_id = $2::uuid)
    )
)`
	var ok bool
	err := r.pool.QueryRow(ctx, q, matchID, reviewerID, revieweeID).Scan(&ok)
	return ok, err
}
