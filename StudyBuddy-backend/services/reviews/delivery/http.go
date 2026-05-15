package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/reviews/domain"
	"studybuddy/backend/services/reviews/usecase"
)

type ReviewsHandler struct {
	Create    usecase.CreateReview
	List      usecase.ListReviewsForUser
	GetRating usecase.GetAverageRating
}

type createReviewRequest struct {
	RevieweeID string `json:"revieweeId"`
	MatchID    string `json:"matchId"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

type reviewResponse struct {
	ID         string `json:"id"`
	ReviewerID string `json:"reviewerId"`
	RevieweeID string `json:"revieweeId"`
	MatchID    string `json:"matchId"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
	CreatedAt  string `json:"createdAt"`
}

type ratingSummaryResponse struct {
	UserID        string  `json:"userId"`
	AverageRating float64 `json:"averageRating"`
	TotalReviews  int     `json:"totalReviews"`
}

func (h *ReviewsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ReviewsHandler) HandleCreateReview(w http.ResponseWriter, r *http.Request) {
	reviewerID := auth.UserIDFromContext(r.Context())
	if reviewerID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	authHdr := r.Header.Get("Authorization")
	out, err := h.Create.CreateReview(r.Context(), usecase.CreateReviewInput{
		ReviewerID: reviewerID,
		RevieweeID: req.RevieweeID,
		MatchID:    req.MatchID,
		Rating:     req.Rating,
		Comment:    req.Comment,
		AuthHeader: authHdr,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSelfReview):
			httputil.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrNoAcceptedMatch):
			httputil.Error(w, http.StatusForbidden, err.Error())
		case errors.Is(err, domain.ErrAlreadyReviewed):
			httputil.Error(w, http.StatusConflict, err.Error())
		case errors.Is(err, domain.ErrInvalidRating), errors.Is(err, domain.ErrCommentTooLong):
			httputil.Error(w, http.StatusBadRequest, err.Error())
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to create review")
		}
		return
	}
	httputil.JSON(w, http.StatusCreated, reviewToJSON(out))
}

func (h *ReviewsHandler) HandleListReviewsForUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	if uid == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	target := chi.URLParam(r, "userID")
	if target == "" {
		httputil.Error(w, http.StatusBadRequest, "missing user id")
		return
	}
	list, err := h.List.ListReviewsForUser(r.Context(), target)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}
	out := make([]reviewResponse, 0, len(list))
	for _, rev := range list {
		out = append(out, reviewToJSON(&rev))
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *ReviewsHandler) HandleListMyReviews(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	if uid == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := h.List.ListReviewsForUser(r.Context(), uid)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}
	out := make([]reviewResponse, 0, len(list))
	for _, rev := range list {
		out = append(out, reviewToJSON(&rev))
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *ReviewsHandler) HandleGetRating(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		httputil.Error(w, http.StatusBadRequest, "missing user id")
		return
	}
	sum, err := h.GetRating.GetAverageRating(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load rating")
		return
	}
	httputil.JSON(w, http.StatusOK, ratingSummaryResponse{
		UserID:        sum.UserID,
		AverageRating: sum.AverageRating,
		TotalReviews:  sum.TotalReviews,
	})
}

func reviewToJSON(r *domain.Review) reviewResponse {
	return reviewResponse{
		ID:         r.ID,
		ReviewerID: r.ReviewerID,
		RevieweeID: r.RevieweeID,
		MatchID:    r.MatchID,
		Rating:     r.Rating,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
