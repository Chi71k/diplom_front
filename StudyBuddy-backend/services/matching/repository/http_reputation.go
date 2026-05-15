package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"studybuddy/backend/services/matching/usecase"
)

// HTTPReputationClient loads a normalized score [0,1] from the reviews service.
type HTTPReputationClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPReputationClient(baseURL string) usecase.ReputationClient {
	return &HTTPReputationClient{
		baseURL: strings.TrimSuffix(strings.TrimSpace(baseURL), "/"),
		client:  &http.Client{Timeout: 4 * time.Second},
	}
}

type ratingDTO struct {
	UserID        string  `json:"userId"`
	AverageRating float64 `json:"averageRating"`
	TotalReviews  int     `json:"totalReviews"`
}

func (c *HTTPReputationClient) GetAverageRating(ctx context.Context, userID string) float64 {
	if c.baseURL == "" {
		return 0.5
	}
	url := fmt.Sprintf("%s/api/v1/reviews/users/%s/rating", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0.5
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0.5
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0.5
	}
	var dto ratingDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return 0.5
	}
	if dto.TotalReviews == 0 {
		return 0.5
	}
	score := dto.AverageRating / 5.0
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}
