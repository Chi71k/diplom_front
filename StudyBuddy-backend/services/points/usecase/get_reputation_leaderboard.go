package usecase

import (
	"context"
	"sort"

	"golang.org/x/sync/errgroup"
)

type ReputationLeaderboardEntry struct {
	UserID        string  `json:"userId"`
	TotalPoints   int64   `json:"totalPoints"`
	AverageRating float64 `json:"averageRating"`
	TotalReviews  int     `json:"totalReviews"`
	Rank          int     `json:"rank"`
}

type GetReputationLeaderboard interface {
	GetReputationLeaderboard(ctx context.Context, limit int) ([]ReputationLeaderboardEntry, error)
}

type getReputationLeaderboard struct {
	repo    PointsRepository
	reviews ReviewRatingReader
}

func NewGetReputationLeaderboard(repo PointsRepository, reviews ReviewRatingReader) GetReputationLeaderboard {
	return &getReputationLeaderboard{repo: repo, reviews: reviews}
}

func (uc *getReputationLeaderboard) GetReputationLeaderboard(ctx context.Context, limit int) ([]ReputationLeaderboardEntry, error) {
	base, err := uc.repo.ListLeaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}
	if len(base) == 0 {
		return nil, nil
	}
	type scored struct {
		ReputationLeaderboardEntry
	}
	out := make([]scored, len(base))
	eg, gCtx := errgroup.WithContext(ctx)
	eg.SetLimit(16)
	for i := range base {
		i := i
		eg.Go(func() error {
			row := base[i]
			avg, n, err := uc.reviews.AverageRating(gCtx, row.UserID)
			if err != nil {
				out[i] = scored{ReputationLeaderboardEntry{
					UserID: row.UserID, TotalPoints: row.TotalPoints, AverageRating: 0, TotalReviews: 0, Rank: row.Rank,
				}}
				return nil
			}
			out[i] = scored{ReputationLeaderboardEntry{
				UserID: row.UserID, TotalPoints: row.TotalPoints, AverageRating: avg, TotalReviews: n, Rank: row.Rank,
			}}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	res := make([]ReputationLeaderboardEntry, len(out))
	for i := range out {
		res[i] = out[i].ReputationLeaderboardEntry
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].AverageRating == res[j].AverageRating {
			return res[i].TotalPoints > res[j].TotalPoints
		}
		return res[i].AverageRating > res[j].AverageRating
	})
	for i := range res {
		res[i].Rank = i + 1
	}
	return res, nil
}
