package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

func NewRouter(h *ReviewsHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)

	r.Get("/api/v1/reviews/users/{userID}/rating", h.HandleGetRating)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Post("/api/v1/reviews", h.HandleCreateReview)
		r.Get("/api/v1/reviews/me", h.HandleListMyReviews)
		r.Get("/api/v1/reviews/users/{userID}", h.HandleListReviewsForUser)
	})

	return r
}
