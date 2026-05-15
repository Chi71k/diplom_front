package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/matching/domain"
	"studybuddy/backend/services/matching/usecase"
)

// PgMatchAccepter accepts a pending match and inserts friendships in one transaction.
type PgMatchAccepter struct {
	pool *pgxpool.Pool
}

func NewPgMatchAccepter(pool *pgxpool.Pool) usecase.MatchAccepter {
	return &PgMatchAccepter{pool: pool}
}

func (a *PgMatchAccepter) AcceptAndBefriend(ctx context.Context, matchID, requesterID, receiverID string) error {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
UPDATE matches
SET status = 'accepted', updated_at = now()
WHERE id = $1 AND status = 'pending';
`, matchID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if tag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		return domain.ErrInvalidStatusChange
	}
	if err := execFriendshipsBothWays(ctx, tx, requesterID, receiverID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
