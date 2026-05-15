package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ReviewsRatingClient struct {
	baseURL string
	client  *http.Client
}

func NewReviewsRatingClient(baseURL string) *ReviewsRatingClient {
	return &ReviewsRatingClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

type ratingSummaryDTO struct {
	UserID        string  `json:"userId"`
	AverageRating float64 `json:"averageRating"`
	TotalReviews  int     `json:"totalReviews"`
}

// AverageRating fetches average rating and count from the reviews HTTP API.
func (c *ReviewsRatingClient) AverageRating(ctx context.Context, userID string) (avg float64, totalReviews int, err error) {
	dto, err := c.fetchRating(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return dto.AverageRating, dto.TotalReviews, nil
}

func (c *ReviewsRatingClient) fetchRating(ctx context.Context, userID string) (ratingSummaryDTO, error) {
	var zero ratingSummaryDTO
	if c == nil || c.baseURL == "" {
		return zero, fmt.Errorf("reviews client not configured")
	}
	url := fmt.Sprintf("%s/api/v1/reviews/users/%s/rating", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return zero, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("reviews: status %d", resp.StatusCode)
	}
	var out ratingSummaryDTO
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return zero, err
	}
	return out, nil
}
