package repository

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	leaderboardRefreshMu   sync.Mutex
	leaderboardLastRefresh time.Time
)

func tryRefreshLeaderboardMV(pool *pgxpool.Pool) {
	leaderboardRefreshMu.Lock()
	defer leaderboardRefreshMu.Unlock()
	if !leaderboardLastRefresh.IsZero() && time.Since(leaderboardLastRefresh) < 5*time.Minute {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, `REFRESH MATERIALIZED VIEW CONCURRENTLY leaderboard_points`); err != nil {
		log.Printf("points: refresh leaderboard mv: %v", err)
		return
	}
	leaderboardLastRefresh = time.Now()
}
