package usecase

import (
	"context"
	"strings"
)

type SearchLeaderboardResult struct {
	UserID        string  `json:"userId"`
	TotalPoints   int64   `json:"totalPoints"`
	Rank          int     `json:"rank"`
	CombinedScore float64 `json:"combinedScore,omitempty"`
}

type SearchLeaderboard interface {
	SearchLeaderboard(ctx context.Context, query string, limit int) ([]SearchLeaderboardResult, error)
}

type searchLeaderboard struct {
	repo     PointsRepository
	fallback GetLeaderboard
}

func NewSearchLeaderboard(repo PointsRepository, fallback GetLeaderboard) SearchLeaderboard {
	return &searchLeaderboard{repo: repo, fallback: fallback}
}

func (uc *searchLeaderboard) SearchLeaderboard(ctx context.Context, query string, limit int) ([]SearchLeaderboardResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return uc.degraded(ctx, limit)
	}

	rows, err := uc.repo.SearchLeaderboardByName(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	out := make([]SearchLeaderboardResult, len(rows))
	for i, row := range rows {
		out[i] = SearchLeaderboardResult{
			UserID:      row.UserID,
			TotalPoints: row.TotalPoints,
			Rank:        i + 1,
		}
	}
	return out, nil
}

func (uc *searchLeaderboard) degraded(ctx context.Context, limit int) ([]SearchLeaderboardResult, error) {
	base, err := uc.fallback.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SearchLeaderboardResult, len(base))
	for i, e := range base {
		out[i] = SearchLeaderboardResult{
			UserID: e.UserID, TotalPoints: e.TotalPoints, Rank: e.Rank, CombinedScore: 0,
		}
	}
	return out, nil
}
