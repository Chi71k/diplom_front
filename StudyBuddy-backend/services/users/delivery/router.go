package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

// NewRouter returns the users service HTTP router.
// JWT secret must match the Auth service secret.
func NewRouter(h *UsersHandler, ih *InterestsHandler, fh *FriendsHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))

		r.Get("/api/v1/users/me", h.HandleGetMe)
		r.Put("/api/v1/users/me", h.HandleUpdateMe)
		r.Delete("/api/v1/users/me", h.HandleDeleteMe)

		r.Get("/api/v1/interests", ih.HandleListCatalog)

		r.Get("/api/v1/users/me/interests", ih.HandleGetMyInterests)
		r.Put("/api/v1/users/me/interests", ih.HandleReplaceMyInterests)

		r.Get("/api/v1/users/me/friends", fh.HandleListFriends)
		r.Delete("/api/v1/users/me/friends/{friendId}", fh.HandleRemoveFriend)
	})

	return r
}
