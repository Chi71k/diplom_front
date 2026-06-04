package delivery

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/users/domain"
	"studybuddy/backend/services/users/usecase"
)

// AdminHandler exposes admin-only HTTP endpoints.
type AdminHandler struct {
	ListUsers     usecase.AdminListUsers
	SetUserActive usecase.AdminSetUserActive
	Stats         usecase.AdminStats
}

type adminUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
}

type adminUserListResponse struct {
	Users    []adminUserResponse `json:"users"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type platformStatsResponse struct {
	TotalUsers    int `json:"totalUsers"`
	ActiveUsers   int `json:"activeUsers"`
	TotalMatches  int `json:"totalMatches"`
	TotalGroups   int `json:"totalGroups"`
	TotalSessions int `json:"totalSessions"`
}

// HandleListUsers GET /api/v1/admin/users
func (h *AdminHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.AdminUserFilter{}
	if role := q.Get("role"); role != "" {
		filter.Role = &role
	}
	if v := q.Get("is_active"); v != "" {
		active, err := strconv.ParseBool(v)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid is_active parameter")
			return
		}
		filter.IsActive = &active
	}

	page := 1
	if p := q.Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	pageSize := 20
	if ps := q.Get("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 {
			pageSize = n
		}
	}

	out, err := h.ListUsers.Execute(r.Context(), filter, page, pageSize)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	users := make([]adminUserResponse, 0, len(out.Users))
	for _, u := range out.Users {
		users = append(users, adminUserResponse{
			ID:        u.ID,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Role:      u.Role,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	httputil.JSON(w, http.StatusOK, adminUserListResponse{
		Users:    users,
		Total:    out.Total,
		Page:     out.Page,
		PageSize: out.PageSize,
	})
}

// HandleDeactivateUser PATCH /api/v1/admin/users/{id}/deactivate
func (h *AdminHandler) HandleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	h.setUserActive(w, r, false)
}

// HandleActivateUser PATCH /api/v1/admin/users/{id}/activate
func (h *AdminHandler) HandleActivateUser(w http.ResponseWriter, r *http.Request) {
	h.setUserActive(w, r, true)
}

func (h *AdminHandler) setUserActive(w http.ResponseWriter, r *http.Request, active bool) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.SetUserActive.Execute(r.Context(), id, active); err != nil {
		if usecase.IsUserNotFound(err) {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandlePlatformStats GET /api/v1/admin/stats
func (h *AdminHandler) HandlePlatformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Stats.Execute(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load platform stats")
		return
	}
	httputil.JSON(w, http.StatusOK, platformStatsResponse{
		TotalUsers:    stats.TotalUsers,
		ActiveUsers:   stats.ActiveUsers,
		TotalMatches:  stats.TotalMatches,
		TotalGroups:   stats.TotalGroups,
		TotalSessions: stats.TotalSessions,
	})
}
