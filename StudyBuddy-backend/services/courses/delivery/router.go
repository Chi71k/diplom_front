package delivery

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
)

// NewRouter returns the courses service HTTP router.
func NewRouter(h *CoursesHandler, jwtSecret []byte) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.HandleHealth)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))

		r.Get("/api/v1/courses", h.HandleListCourses)
		r.Post("/api/v1/courses", h.HandleCreateCourse)
		r.Get("/api/v1/courses/{courseId}", h.HandleGetCourse)
		r.Patch("/api/v1/courses/{courseId}", h.HandlePatchCourse)
		r.Delete("/api/v1/courses/{courseId}", h.HandleDeleteCourse)
	})

	return r
}
