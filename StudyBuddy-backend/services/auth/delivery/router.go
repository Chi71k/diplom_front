package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter returns the auth service HTTP router.
func NewRouter(h *AuthHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)
	r.Post("/api/v1/auth/register", h.HandleRegister)
	r.Post("/api/v1/auth/login", h.HandleLogin)
	r.Post("/api/v1/auth/refresh", h.HandleRefresh)

	return r
}
