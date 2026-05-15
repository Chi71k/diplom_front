package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

func NewRouter(h *PointsHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)
	r.Get("/api/v1/users/{userID}/points", h.HandleGetPublicUserPoints)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Post("/api/v1/points/events", h.HandlePostEvent)
		r.Get("/api/v1/points/me", h.HandleGetMine)
		r.Get("/api/v1/points/leaderboard/search", h.HandleSearchLeaderboard)
		r.Get("/api/v1/points/leaderboard/reputation", h.HandleGetReputationLeaderboard)
		r.Get("/api/v1/points/leaderboard", h.HandleGetLeaderboard)
	})

	return r
}
