package delivery

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/users/domain"
	"studybuddy/backend/services/users/usecase"
)

var userPathUUID = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// UsersHandler exposes user profile HTTP endpoints.
type UsersHandler struct {
	GetMe    usecase.GetMe
	GetUser  usecase.GetUser
	UpdateMe usecase.UpdateMe
	DeleteMe usecase.DeleteMe
}

// UserProfileResponse matches OpenAPI UserProfile (minimal).
type UserProfileResponse struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	FirstName   string            `json:"firstName"`
	LastName    string            `json:"lastName"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatarUrl"`
	TelegramTag string            `json:"telegramTag"`
	Interests   []domain.Interest `json:"interests"`
}

// PublicUserProfileResponse is the same as UserProfileResponse but without email.
type PublicUserProfileResponse struct {
	ID          string            `json:"id"`
	FirstName   string            `json:"firstName"`
	LastName    string            `json:"lastName"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatarUrl"`
	TelegramTag string            `json:"telegramTag"`
	Interests   []domain.Interest `json:"interests"`
}

// UpdateProfileRequest matches OpenAPI UpdateProfileRequest.
type UpdateProfileRequest struct {
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	TelegramTag *string `json:"telegramTag,omitempty"`
}

// HandleGetMe GET /api/v1/users/me (requires JWT).
func (h *UsersHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	profile, err := h.GetMe.GetMe(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get profile")
		return
	}
	if profile == nil {
		httputil.Error(w, http.StatusNotFound, "profile not found")
		return
	}
	httputil.JSON(w, http.StatusOK, profileToResponse(profile))
}

// HandleGetUser GET /api/v1/users/{id} (requires JWT).
func (h *UsersHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	if auth.UserIDFromContext(r.Context()) == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" || !userPathUUID.MatchString(id) {
		httputil.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	profile, err := h.GetUser.GetUser(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get profile")
		return
	}
	if profile == nil {
		httputil.Error(w, http.StatusNotFound, "profile not found")
		return
	}
	httputil.JSON(w, http.StatusOK, profileToPublicResponse(profile))
}

// HandleUpdateMe PUT /api/v1/users/me (requires JWT).
func (h *UsersHandler) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	_, _ = io.Copy(io.Discard, r.Body)
	in := usecase.UpdateMeInput{UserID: userID}
	in.FirstName = req.FirstName
	in.LastName = req.LastName
	in.Bio = req.Bio
	in.AvatarURL = req.AvatarURL
	in.TelegramTag = req.TelegramTag
	profile, err := h.UpdateMe.UpdateMe(r.Context(), in)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	httputil.JSON(w, http.StatusOK, profileToResponse(profile))
}

// HandleDeleteMe DELETE /api/v1/users/me (requires JWT).
func (h *UsersHandler) HandleDeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.DeleteMe.DeleteMe(r.Context(), userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete account")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleHealth GET /health
func (h *UsersHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func profileToResponse(p *domain.Profile) UserProfileResponse {
	return UserProfileResponse{
		ID:          p.UserID,
		Email:       p.Email,
		FirstName:   p.FirstName,
		LastName:    p.LastName,
		Bio:         p.Bio,
		AvatarURL:   p.AvatarURL,
		TelegramTag: p.TelegramTag,
		Interests:   nil, // TODO: load when interests endpoint is added
	}
}

func profileToPublicResponse(p *domain.Profile) PublicUserProfileResponse {
	return PublicUserProfileResponse{
		ID:          p.UserID,
		FirstName:   p.FirstName,
		LastName:    p.LastName,
		Bio:         p.Bio,
		AvatarURL:   p.AvatarURL,
		TelegramTag: p.TelegramTag,
		Interests:   nil, // TODO: load when interests endpoint is added
	}
}
