package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/courses/domain"
	"studybuddy/backend/services/courses/usecase"
)

// CoursesHandler exposes courses HTTP endpoints.
type CoursesHandler struct {
	List   usecase.ListCourses
	Get    usecase.GetCourse
	Create usecase.CreateCourse
	Update usecase.UpdateCourse
	Delete usecase.DeleteCourse
}

// CourseResponse is the API shape for a course.
type CourseResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Level       string `json:"level"`
	OwnerUserID string `json:"ownerUserId"`
}

// CreateCourseRequest is the body for creating a course.
type CreateCourseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Level       string `json:"level"`
}

// UpdateCourseRequest is the body for partially updating a course.
type UpdateCourseRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Subject     *string `json:"subject"`
	Level       *string `json:"level"`
}

// HandleHealth GET /health
func (h *CoursesHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleListCourses GET /api/v1/courses
func (h *CoursesHandler) HandleListCourses(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := usecase.ListCoursesFilter{
		Subject: q.Get("subject"),
		Level:   q.Get("level"),
		Limit:   20,
		Offset:  0,
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filter.Offset = n
		}
	}
	courses, err := h.List.List(r.Context(), filter)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list courses")
		return
	}
	resp := make([]CourseResponse, 0, len(courses))
	for _, c := range courses {
		resp = append(resp, courseToResponse(&c))
	}
	httputil.JSON(w, http.StatusOK, resp)
}

// HandleGetCourse GET /api/v1/courses/{courseId}
func (h *CoursesHandler) HandleGetCourse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "courseId")
	h.handleGetCourse(w, r, id)
}

func (h *CoursesHandler) handleGetCourse(w http.ResponseWriter, r *http.Request, id string) {
	if _, err := uuid.Parse(id); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid course id")
		return
	}
	course, err := h.Get.Get(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get course")
		return
	}
	if course == nil {
		httputil.Error(w, http.StatusNotFound, "course not found")
		return
	}
	httputil.JSON(w, http.StatusOK, courseToResponse(course))
}

// HandleCreateCourse POST /api/v1/courses
func (h *CoursesHandler) HandleCreateCourse(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Title == "" || req.Description == "" || req.Subject == "" || req.Level == "" {
		httputil.Error(w, http.StatusBadRequest, "title, description, subject, level required")
		return
	}
	course, err := h.Create.Create(r.Context(), usecase.CreateCourseInput{
		Title:       req.Title,
		Description: req.Description,
		Subject:     req.Subject,
		Level:       req.Level,
		OwnerUserID: userID,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create course")
		return
	}
	httputil.JSON(w, http.StatusCreated, courseToResponse(course))
}

// HandlePatchCourse PATCH /api/v1/courses/{courseId}
func (h *CoursesHandler) HandlePatchCourse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "courseId")
	h.handleUpdateCourse(w, r, id)
}

func (h *CoursesHandler) handleUpdateCourse(w http.ResponseWriter, r *http.Request, id string) {
	if _, err := uuid.Parse(id); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid course id")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Title == nil && req.Description == nil && req.Subject == nil && req.Level == nil {
		course, err := h.Get.Get(r.Context(), id)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "failed to get course")
			return
		}
		if course == nil {
			httputil.Error(w, http.StatusNotFound, "course not found")
			return
		}
		if course.OwnerUserID != userID {
			httputil.Error(w, http.StatusForbidden, "you do not own this course")
			return
		}
		httputil.JSON(w, http.StatusOK, courseToResponse(course))
		return
	}
	course, err := h.Update.Update(r.Context(), usecase.UpdateCourseInput{
		ID:             id,
		RequestingUser: userID,
		Title:          req.Title,
		Description:    req.Description,
		Subject:        req.Subject,
		Level:          req.Level,
	})
	if err != nil {
		switch err {
		case domain.ErrForbidden:
			httputil.Error(w, http.StatusForbidden, "you do not own this course")
		case domain.ErrCourseNotFound:
			httputil.Error(w, http.StatusNotFound, "course not found")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to update course")
		}
		return
	}
	httputil.JSON(w, http.StatusOK, courseToResponse(course))
}

// HandleDeleteCourse DELETE /api/v1/courses/{courseId}
func (h *CoursesHandler) HandleDeleteCourse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "courseId")
	h.handleDeleteCourse(w, r, id)
}

func (h *CoursesHandler) handleDeleteCourse(w http.ResponseWriter, r *http.Request, id string) {
	if _, err := uuid.Parse(id); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid course id")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.Delete.Delete(r.Context(), usecase.DeleteCourseInput{
		ID:             id,
		RequestingUser: userID,
	}); err != nil {
		switch err {
		case domain.ErrForbidden:
			httputil.Error(w, http.StatusForbidden, "you do not own this course")
		case domain.ErrCourseNotFound:
			httputil.Error(w, http.StatusNotFound, "course not found")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to delete course")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func courseToResponse(c *domain.Course) CourseResponse {
	return CourseResponse{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		Subject:     c.Subject,
		Level:       c.Level,
		OwnerUserID: c.OwnerUserID,
	}
}
