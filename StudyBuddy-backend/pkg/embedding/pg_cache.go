package embedding

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgCache stores embeddings in PostgreSQL (float8[]).
type PgCache struct {
	pool *pgxpool.Pool
}

func NewPgCache(pool *pgxpool.Pool) *PgCache {
	return &PgCache{pool: pool}
}

func (c *PgCache) Get(ctx context.Context, userID string) ([]float64, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT embedding, profile_hash
FROM user_embedding_cache
WHERE user_id = $1;
`
	var (
		embedding []float64
		hash      string
	)
	err := c.pool.QueryRow(ctx, q, userID).Scan(&embedding, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("embedding cache get: %w", err)
	}
	return embedding, hash, nil
}

func (c *PgCache) Set(ctx context.Context, userID string, embedding []float64, profileHash string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
INSERT INTO user_embedding_cache (user_id, embedding, profile_hash, updated_at)
VALUES ($1, $2::float8[], $3, now())
ON CONFLICT (user_id) DO UPDATE SET
  embedding = EXCLUDED.embedding,
  profile_hash = EXCLUDED.profile_hash,
  updated_at = now();
`
	_, err := c.pool.Exec(ctx, q, userID, embedding, profileHash)
	if err != nil {
		return fmt.Errorf("embedding cache set: %w", err)
	}
	return nil
}

func (c *PgCache) Delete(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `DELETE FROM user_embedding_cache WHERE user_id = $1`
	_, err := c.pool.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("embedding cache delete: %w", err)
	}
	return nil
}
