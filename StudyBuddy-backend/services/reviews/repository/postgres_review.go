package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/reviews/domain"
)

type PostgresReviewRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresReviewRepository(pool *pgxpool.Pool) *PostgresReviewRepository {
	return &PostgresReviewRepository{pool: pool}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *PostgresReviewRepository) Create(ctx context.Context, reviewerID, revieweeID, matchID string, rating int, comment string) (*domain.Review, error) {
	const q = `
INSERT INTO reviews (reviewer_id, reviewee_id, match_id, rating, comment)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5)
RETURNING id::text, reviewer_id::text, reviewee_id::text, match_id::text, rating, comment, created_at
`
	row := r.pool.QueryRow(ctx, q, reviewerID, revieweeID, matchID, rating, comment)
	var rev domain.Review
	if err := row.Scan(&rev.ID, &rev.ReviewerID, &rev.RevieweeID, &rev.MatchID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrAlreadyReviewed
		}
		return nil, err
	}
	return &rev, nil
}

func (r *PostgresReviewRepository) ListForReviewee(ctx context.Context, revieweeID string) ([]domain.Review, error) {
	const q = `
SELECT id::text, reviewer_id::text, reviewee_id::text, match_id::text, rating, comment, created_at
FROM reviews
WHERE reviewee_id = $1::uuid
ORDER BY created_at DESC
`
	rows, err := r.pool.Query(ctx, q, revieweeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var rev domain.Review
		if err := rows.Scan(&rev.ID, &rev.ReviewerID, &rev.RevieweeID, &rev.MatchID, &rev.Rating, &rev.Comment, &rev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	return out, rows.Err()
}

func (r *PostgresReviewRepository) GetAverageRating(ctx context.Context, revieweeID string) (domain.RatingSummary, error) {
	const q = `
SELECT COALESCE(AVG(rating)::float8, 0), COUNT(*)::int
FROM reviews
WHERE reviewee_id = $1::uuid
`
	var sum domain.RatingSummary
	sum.UserID = revieweeID
	err := r.pool.QueryRow(ctx, q, revieweeID).Scan(&sum.AverageRating, &sum.TotalReviews)
	sum.UserID = revieweeID
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sum, nil
		}
		return sum, err
	}
	return sum, nil
}
