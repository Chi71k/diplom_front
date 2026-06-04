package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/matching/domain"
	"studybuddy/backend/services/matching/usecase"
)

var uuidRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type MatchingHandler struct {
	ListCandidates usecase.ListCandidates
	SendRequest    usecase.SendMatchRequest
	Respond        usecase.RespondToMatch
	Cancel         usecase.CancelMatch
	ListMatches    usecase.ListMatches
	Profiles       usecase.ProfileClient
}

type MatchResponse struct {
	ID          string `json:"id"`
	RequesterID string `json:"requesterId"`
	ReceiverID  string `json:"receiverId"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// MatchResponseEnriched adds display names for list responses.
type MatchResponseEnriched struct {
	MatchResponse
	RequesterName string `json:"requesterName"`
	ReceiverName  string `json:"receiverName"`
}

type SlotOverlapResponse struct {
	DayOfWeek int    `json:"dayOfWeek"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Timezone  string `json:"timezone"`
}

type CandidateResponse struct {
	UserID             string                `json:"userId"`
	FirstName          string                `json:"firstName"`
	LastName           string                `json:"lastName"`
	Bio                string                `json:"bio"`
	AvatarURL          string                `json:"avatarUrl"`
	CommonCourses      []string              `json:"commonCourses"`
	CommonSlots        []SlotOverlapResponse `json:"commonSlots"`
	SemanticScore      float64               `json:"semanticScore"`
	AvailabilityScore  float64               `json:"availabilityScore"`
	CourseScore        float64               `json:"courseScore"`
	ReputationScore    float64               `json:"reputationScore"`
	MutualFriendsScore float64               `json:"mutualFriendsScore"`
	OverallScore       float64               `json:"overallScore"`
}

type SendMatchRequestBody struct {
	ReceiverID string `json:"receiverId"`
	Message    string `json:"message"`
}

type RespondMatchBody struct {
	Accept bool `json:"accept"`
}

// health checks
func (h *MatchingHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/v1/matching/candidates

func (h *MatchingHandler) HandleListCandidates(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	candidates, err := h.ListCandidates.ListCandidates(r.Context(), usecase.ListCandidatesInput{
		RequesterID: userID,
		Limit:       limit,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list candidates")
		return
	}

	resp := make([]CandidateResponse, 0, len(candidates))
	for _, c := range candidates {
		resp = append(resp, candidateToResponse(c))
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": resp})
}

// POST /api/v1/matching/requests

func (h *MatchingHandler) HandleSendRequest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body SendMatchRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, _ = io.Copy(io.Discard, r.Body)
	if body.ReceiverID == "" {
		httputil.Error(w, http.StatusBadRequest, "receiverId is required")
		return
	}
	if !uuidRE.MatchString(body.ReceiverID) {
		httputil.Error(w, http.StatusBadRequest, "receiverId must be a valid uuid")
		return
	}

	m, err := h.SendRequest.Send(r.Context(), usecase.SendMatchRequestInput{
		RequesterID: userID,
		ReceiverID:  body.ReceiverID,
		Message:     body.Message,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCannotMatchSelf):
			httputil.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrMatchAlreadyExists):
			httputil.Error(w, http.StatusConflict, err.Error())
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to send match request")
		}
		return
	}
	httputil.JSON(w, http.StatusCreated, matchToResponse(m))
}

// GET /api/v1/matching/requests
func (h *MatchingHandler) HandleListMatches(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	status := domain.MatchStatus(q.Get("status"))

	limit := 20
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	matches, err := h.ListMatches.List(r.Context(), usecase.ListMatchesInput{
		UserID: userID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list match requests")
		return
	}

	resp := make([]MatchResponseEnriched, 0, len(matches))
	nameCache := make(map[string]string)
	for i := range matches {
		mr := matchToResponse(&matches[i])
		resp = append(resp, MatchResponseEnriched{
			MatchResponse: mr,
			RequesterName: h.profileDisplayName(r.Context(), matches[i].RequesterID, nameCache),
			ReceiverName:  h.profileDisplayName(r.Context(), matches[i].ReceiverID, nameCache),
		})
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": resp})
}

// POST /api/v1/matching/requests/{id}/respond
func (h *MatchingHandler) HandleRespond(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	matchID := chi.URLParam(r, "id")
	if matchID == "" || !uuidRE.MatchString(matchID) {
		httputil.Error(w, http.StatusNotFound, "not found")
		return
	}

	var body RespondMatchBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, _ = io.Copy(io.Discard, r.Body)

	m, err := h.Respond.Respond(r.Context(), usecase.RespondToMatchInput{
		MatchID:     matchID,
		ResponderID: userID,
		Accept:      body.Accept,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMatchNotFound):
			httputil.Error(w, http.StatusNotFound, "match not found")
		case errors.Is(err, domain.ErrForbidden):
			httputil.Error(w, http.StatusForbidden, "you are not the receiver of this request")
		case errors.Is(err, domain.ErrInvalidStatusChange):
			httputil.Error(w, http.StatusConflict, "match is not in pending state")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to respond to match request")
		}
		return
	}
	httputil.JSON(w, http.StatusOK, matchToResponse(m))
}

// DELETE /api/v1/matching/requests/{id}
func (h *MatchingHandler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	matchID := chi.URLParam(r, "id")
	if matchID == "" || !uuidRE.MatchString(matchID) {
		httputil.Error(w, http.StatusNotFound, "not found")
		return
	}

	if err := h.Cancel.Cancel(r.Context(), usecase.CancelMatchInput{
		MatchID:     matchID,
		RequesterID: userID,
	}); err != nil {
		switch {
		case errors.Is(err, domain.ErrMatchNotFound):
			httputil.Error(w, http.StatusNotFound, "match not found")
		case errors.Is(err, domain.ErrForbidden):
			httputil.Error(w, http.StatusForbidden, "you did not send this request")
		case errors.Is(err, domain.ErrInvalidStatusChange):
			httputil.Error(w, http.StatusConflict, "only pending requests can be cancelled")
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to cancel match request")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helpers
func (h *MatchingHandler) profileDisplayName(ctx context.Context, userID string, cache map[string]string) string {
	if n, ok := cache[userID]; ok {
		return n
	}
	var name string
	if h.Profiles != nil {
		p, err := h.Profiles.GetProfile(ctx, userID)
		if err == nil && p != nil {
			name = strings.TrimSpace(p.FirstName + " " + p.LastName)
		}
	}
	cache[userID] = name
	return name
}

func matchToResponse(m *domain.Match) MatchResponse {
	return MatchResponse{
		ID:          m.ID,
		RequesterID: m.RequesterID,
		ReceiverID:  m.ReceiverID,
		Status:      string(m.Status),
		Message:     m.Message,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func candidateToResponse(c domain.MatchCandidate) CandidateResponse {
	slots := make([]SlotOverlapResponse, 0, len(c.CommonSlots))
	for _, s := range c.CommonSlots {
		slots = append(slots, SlotOverlapResponse{
			DayOfWeek: s.DayOfWeek,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Timezone:  s.Timezone,
		})
	}
	courses := c.CommonCourses
	if courses == nil {
		courses = []string{}
	}
	return CandidateResponse{
		UserID:             c.UserID,
		FirstName:          c.FirstName,
		LastName:           c.LastName,
		Bio:                c.Bio,
		AvatarURL:          c.AvatarURL,
		CommonCourses:      courses,
		CommonSlots:        slots,
		SemanticScore:      c.SemanticScore,
		AvailabilityScore:  c.AvailScore,
		CourseScore:        c.CourseScore,
		ReputationScore:    c.ReputationScore,
		MutualFriendsScore: c.MutualFriendsScore,
		OverallScore:       c.OverallScore,
	}
}
