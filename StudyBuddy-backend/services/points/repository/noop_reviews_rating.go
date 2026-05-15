package repository

import "context"

// NoopReviewsRating returns empty ratings when the reviews service URL is not configured.
type NoopReviewsRating struct{}

func (NoopReviewsRating) AverageRating(ctx context.Context, userID string) (float64, int, error) {
	_ = ctx
	_ = userID
	return 0, 0, nil
}
