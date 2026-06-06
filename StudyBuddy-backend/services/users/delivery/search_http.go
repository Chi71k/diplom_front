package delivery

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/users/usecase"
)

type StudentSearchResponse struct {
	UserID     string  `json:"userId"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Bio        string  `json:"bio"`
	AvatarURL  string  `json:"avatarUrl"`
	Similarity float64 `json:"similarity"`
}

// HandleSearchStudents GET /api/v1/users/search?q=...&limit=N
func (h *UsersHandler) HandleSearchStudents(w http.ResponseWriter, r *http.Request) {
	searcherID := auth.UserIDFromContext(r.Context())
	if searcherID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		httputil.Error(w, http.StatusBadRequest, "q is required")
		return
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	hits, err := h.SearchStudents.SearchStudents(r.Context(), usecase.SearchStudentsInput{
		SearcherID: searcherID,
		Query:      query,
		Limit:      limit,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrSearchQueryRequired) {
			httputil.Error(w, http.StatusBadRequest, "q is required")
			return
		}
		log.Printf("search students error: %v", err)
		httputil.Error(w, http.StatusInternalServerError, "failed to search students")
		return
	}

	resp := make([]StudentSearchResponse, 0, len(hits))
	for _, hit := range hits {
		resp = append(resp, StudentSearchResponse{
			UserID:     hit.UserID,
			FirstName:  hit.FirstName,
			LastName:   hit.LastName,
			Bio:        hit.Bio,
			AvatarURL:  hit.AvatarURL,
			Similarity: hit.Similarity,
		})
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": resp})
}
