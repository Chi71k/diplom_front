package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

func NewRouter(h *GroupsHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))

		r.Post("/api/v1/groups", h.HandleCreateGroup)
		r.Get("/api/v1/groups", h.HandleListMyGroups)
		r.Get("/api/v1/groups/{groupID}", h.HandleGetGroup)
		r.Patch("/api/v1/groups/{groupID}", h.HandleUpdateGroup)
		r.Delete("/api/v1/groups/{groupID}", h.HandleDeleteGroup)
		r.Post("/api/v1/groups/{groupID}/members", h.HandleInviteMember)
		r.Delete("/api/v1/groups/{groupID}/members/{userID}", h.HandleRemoveMember)
		r.Get("/api/v1/groups/{groupID}/suggestions", h.HandleSuggestions)
	})

	return r
}
