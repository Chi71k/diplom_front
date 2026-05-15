package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

func NewRouter(h *MatchingHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))

		r.Get("/api/v1/matching/candidates", h.HandleListCandidates)
		r.Post("/api/v1/matching/requests", h.HandleSendRequest)
		r.Get("/api/v1/matching/requests", h.HandleListMatches)
		r.Post("/api/v1/matching/requests/{id}/respond", h.HandleRespond)
		r.Delete("/api/v1/matching/requests/{id}", h.HandleCancel)
	})

	return r
}
