package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/groups/domain"
	"studybuddy/backend/services/groups/usecase"
)

type GroupsHandler struct {
	Create   usecase.CreateGroup
	Get      usecase.GetGroup
	ListMine usecase.ListMyGroups
	Invite   usecase.InviteMember
	Remove   usecase.RemoveMember
	Delete   usecase.DeleteGroup
	Update   usecase.UpdateGroup
	Suggest  usecase.SuggestMembersForGroup
}

func (h *GroupsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createGroupBody struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CourseIDs   []string `json:"courseIds"`
}

func (h *GroupsHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body createGroupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	g, err := h.Create.CreateGroup(r.Context(), usecase.CreateGroupInput{
		OwnerID:     userID,
		Name:        body.Name,
		Description: body.Description,
		CourseIDs:   body.CourseIDs,
	})
	if err != nil {
		mapGroupErr(w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, groupToJSON(g))
}

func (h *GroupsHandler) HandleListMyGroups(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.ListMine.ListMyGroups(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list groups")
		return
	}
	out := make([]any, 0, len(list))
	for i := range list {
		out = append(out, groupToJSON(&list[i]))
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *GroupsHandler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "groupID")
	if _, err := uuid.Parse(id); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group id")
		return
	}
	g, err := h.Get.GetGroup(r.Context(), id)
	if err != nil {
		mapGroupErr(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, groupToJSON(g))
}

type patchGroupBody struct {
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	CourseIDs   *[]string `json:"courseIds"`
}

func (h *GroupsHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := chi.URLParam(r, "groupID")
	if _, err := uuid.Parse(groupID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var body patchGroupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	g, err := h.Update.UpdateGroup(r.Context(), usecase.UpdateGroupInput{
		ActorID:     userID,
		GroupID:     groupID,
		Name:        body.Name,
		Description: body.Description,
		CourseIDs:   body.CourseIDs,
	})
	if err != nil {
		mapGroupErr(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, groupToJSON(g))
}

func (h *GroupsHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := chi.URLParam(r, "groupID")
	if _, err := uuid.Parse(groupID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group id")
		return
	}
	if err := h.Delete.DeleteGroup(r.Context(), usecase.DeleteGroupInput{ActorID: userID, GroupID: groupID}); err != nil {
		mapGroupErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type inviteBody struct {
	UserID string `json:"userId"`
}

func (h *GroupsHandler) HandleInviteMember(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := chi.URLParam(r, "groupID")
	if _, err := uuid.Parse(groupID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var body inviteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
		httputil.Error(w, http.StatusBadRequest, "userId required")
		return
	}
	if _, err := uuid.Parse(body.UserID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	err := h.Invite.InviteMember(r.Context(), usecase.InviteMemberInput{
		ActorID:   userID,
		GroupID:   groupID,
		InviteeID: body.UserID,
	})
	if err != nil {
		mapGroupErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupsHandler) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	groupID := chi.URLParam(r, "groupID")
	target := chi.URLParam(r, "userID")
	if _, err := uuid.Parse(groupID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group id")
		return
	}
	if _, err := uuid.Parse(target); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	err := h.Remove.RemoveMember(r.Context(), usecase.RemoveMemberInput{
		ActorID:  userID,
		GroupID:  groupID,
		TargetID: target,
	})
	if err != nil {
		mapGroupErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupsHandler) HandleSuggestions(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")
	if _, err := uuid.Parse(groupID); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid group id")
		return
	}
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	items, err := h.Suggest.SuggestMembersForGroup(r.Context(), usecase.SuggestMembersForGroupInput{
		GroupID: groupID,
		Limit:   limit,
	})
	if err != nil {
		mapGroupErr(w, err)
		return
	}
	resp := make([]map[string]any, 0, len(items))
	for _, it := range items {
		resp = append(resp, map[string]any{
			"userId":          it.UserID,
			"firstName":       it.FirstName,
			"lastName":        it.LastName,
			"avatarUrl":       it.AvatarURL,
			"similarityScore": it.SimilarityScore,
		})
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": resp})
}

func groupToJSON(g *domain.Group) map[string]any {
	members := make([]map[string]any, 0, len(g.Members))
	for _, m := range g.Members {
		members = append(members, map[string]any{
			"userId":   m.UserID,
			"role":     string(m.Role),
			"joinedAt": m.JoinedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	courses := g.CourseIDs
	if courses == nil {
		courses = []string{}
	}
	return map[string]any{
		"id":          g.ID,
		"name":        g.Name,
		"description": g.Description,
		"ownerId":     g.OwnerID,
		"courseIds":   courses,
		"members":     members,
		"createdAt":   g.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func mapGroupErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrGroupNotFound):
		httputil.Error(w, http.StatusNotFound, "group not found")
	case errors.Is(err, domain.ErrGroupFull):
		httputil.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrAlreadyMember):
		httputil.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrNotMember):
		httputil.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrNotOwner):
		httputil.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		httputil.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrOwnerCannotLeave):
		httputil.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidGroupName):
		httputil.Error(w, http.StatusBadRequest, err.Error())
	default:
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "23503") {
			httputil.Error(w, http.StatusBadRequest, "invalid course reference")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "request failed")
	}
}
