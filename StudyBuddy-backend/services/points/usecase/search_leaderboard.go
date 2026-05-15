package usecase

import (
	"context"
	"math"
	"sort"
	"strings"

	"studybuddy/backend/pkg/embedding"
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
	embed    *embedding.Client
	embProv  EmbeddingProvider
	fallback GetLeaderboard
}

func NewSearchLeaderboard(repo PointsRepository, embed *embedding.Client, embProv EmbeddingProvider, fallback GetLeaderboard) SearchLeaderboard {
	return &searchLeaderboard{repo: repo, embed: embed, embProv: embProv, fallback: fallback}
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
	qVec, err := uc.embed.Embed(ctx, q)
	if err != nil || qVec == nil || len(qVec) == 0 {
		return uc.degraded(ctx, limit)
	}
	rows, err := uc.repo.ListAllLeaderboardTotals(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	maxPts, err := uc.repo.MaxLeaderboardPoints(ctx)
	if err != nil {
		return nil, err
	}
	if maxPts <= 0 {
		maxPts = 1
	}
	type rowScore struct {
		UserID      string
		TotalPoints int64
		Rank        int
		Score       float64
	}
	var scored []rowScore
	for _, row := range rows {
		uvec, err := uc.embProv.GetOrCompute(ctx, row.UserID)
		if err != nil || len(uvec) == 0 {
			continue
		}
		cos := embedding.CosineSimilarity(qVec, uvec)
		if math.IsNaN(cos) {
			cos = 0
		}
		norm := float64(row.TotalPoints) / float64(maxPts)
		if math.IsNaN(norm) {
			norm = 0
		}
		s := 0.6*cos + 0.4*norm
		scored = append(scored, rowScore{
			UserID: row.UserID, TotalPoints: row.TotalPoints, Rank: row.Rank, Score: s,
		})
	}
	if len(scored) == 0 {
		return uc.degraded(ctx, limit)
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].TotalPoints > scored[j].TotalPoints
		}
		return scored[i].Score > scored[j].Score
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}
	out := make([]SearchLeaderboardResult, len(scored))
	for i, s := range scored {
		out[i] = SearchLeaderboardResult{
			UserID: s.UserID, TotalPoints: s.TotalPoints, Rank: i + 1, CombinedScore: s.Score,
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
