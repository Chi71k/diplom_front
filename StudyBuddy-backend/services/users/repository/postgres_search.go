package repository

import (
	"context"
	"fmt"
	"time"

	"studybuddy/backend/services/users/usecase"
)

func (r *PgProfileRepository) SearchByEmbedding(ctx context.Context, excludeUserID string, vector []float32, limit int) ([]usecase.StudentSearchHit, error) {
	if len(vector) != 768 {
		return nil, fmt.Errorf("embedding must have 768 dimensions, got %d", len(vector))
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	const q = `
SELECT id::text, first_name, last_name, COALESCE(bio, ''), COALESCE(avatar_url, ''),
       1 - (embedding_vector <=> $1::vector) AS similarity
FROM users
WHERE is_active = true
  AND role != 'admin'
  AND embedding_vector IS NOT NULL
  AND id <> $2::uuid
ORDER BY embedding_vector <=> $1::vector
LIMIT $3;
`
	rows, err := r.pool.Query(ctx, q, formatVector(vector), excludeUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("search by embedding: %w", err)
	}
	defer rows.Close()

	var out []usecase.StudentSearchHit
	for rows.Next() {
		var hit usecase.StudentSearchHit
		if err := rows.Scan(&hit.UserID, &hit.FirstName, &hit.LastName, &hit.Bio, &hit.AvatarURL, &hit.Similarity); err != nil {
			return nil, err
		}
		out = append(out, hit)
	}
	return out, rows.Err()
}
