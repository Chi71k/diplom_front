package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

func NewRouter(h *AvailabilityHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)

	// Google OAuth callback: browser redirect, no JWT.
	r.Get("/api/v1/availability/gcal/callback", h.HandleGCalCallback)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))

		r.Get("/api/v1/availability/slots", h.HandleListSlots)
		r.Post("/api/v1/availability/slots", h.HandleCreateSlot)
		r.Delete("/api/v1/availability/slots/{slotId}", h.HandleDeleteSlot)

		r.Get("/api/v1/availability/gcal/connect", h.HandleGCalConnect)
		r.Post("/api/v1/availability/gcal/import", h.HandleGCalImport)
		r.Delete("/api/v1/availability/gcal/disconnect", h.HandleGCalDisconnect)

		r.Post("/api/v1/sessions", h.HandleProposeSession)
		r.Get("/api/v1/sessions", h.HandleListMySessions)
		r.Get("/api/v1/sessions/{sessionID}", h.HandleGetSession)
		r.Post("/api/v1/sessions/{sessionID}/confirm", h.HandleConfirmSession)
		r.Delete("/api/v1/sessions/{sessionID}", h.HandleCancelSession)
	})

	return r
}
