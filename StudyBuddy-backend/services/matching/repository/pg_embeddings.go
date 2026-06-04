package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/matching/usecase"
)

// PgEmbeddingStore reads precomputed embedding vectors from the users table.
type PgEmbeddingStore struct {
	pool *pgxpool.Pool
}

// NewPgEmbeddingStore creates an embedding store backed by PostgreSQL.
func NewPgEmbeddingStore(pool *pgxpool.Pool) usecase.EmbeddingStore {
	return &PgEmbeddingStore{pool: pool}
}

// GetEmbeddings returns stored vectors for the given user IDs (missing users omitted).
func (s *PgEmbeddingStore) GetEmbeddings(ctx context.Context, userIDs []string) (map[string][]float32, error) {
	if len(userIDs) == 0 {
		return map[string][]float32{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT id::text, embedding_vector::text
FROM users
WHERE id = ANY($1::uuid[])
  AND embedding_vector IS NOT NULL;
`
	rows, err := s.pool.Query(ctx, q, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get embeddings: %w", err)
	}
	defer rows.Close()

	out := make(map[string][]float32, len(userIDs))
	for rows.Next() {
		var id string
		var raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		vec, err := parseEmbeddingVector(raw)
		if err != nil {
			return nil, err
		}
		if len(vec) > 0 {
			out[id] = vec
		}
	}
	return out, rows.Err()
}

func parseEmbeddingVector(s string) ([]float32, error) {
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
