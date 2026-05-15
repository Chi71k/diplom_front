package delivery

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	mdomain "studybuddy/backend/services/matching/domain"
	"studybuddy/backend/services/users/usecase"
)

var friendUUIDRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type FriendsHandler struct {
	List   usecase.ListFriends
	Remove usecase.RemoveFriend
}

// GET /api/v1/users/me/friends
func (h *FriendsHandler) HandleListFriends(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.List.ListFriends(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list friends")
		return
	}
	if items == nil {
		items = []string{}
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// DELETE /api/v1/users/me/friends/{friendId}
func (h *FriendsHandler) HandleRemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	friendID := chi.URLParam(r, "friendId")
	if friendID == "" || !friendUUIDRE.MatchString(friendID) {
		httputil.Error(w, http.StatusBadRequest, "invalid friend id")
		return
	}
	if err := h.Remove.RemoveFriend(r.Context(), userID, friendID); err != nil {
		if errors.Is(err, mdomain.ErrFriendshipNotFound) {
			httputil.Error(w, http.StatusNotFound, "friendship not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to remove friend")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
