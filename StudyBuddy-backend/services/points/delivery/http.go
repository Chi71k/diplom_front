package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/pkg/httputil"
	"studybuddy/backend/services/points/domain"
	"studybuddy/backend/services/points/usecase"
)

type PointsHandler struct {
	AddTx          usecase.AddTransaction
	GetMine        usecase.GetMyPoints
	GetBoard       usecase.GetLeaderboard
	GetRepBoard    usecase.GetReputationLeaderboard
	SearchBoard    usecase.SearchLeaderboard
	GetPublicTotal usecase.GetUserPointsTotal
}

type postPointsEventRequest struct {
	UserID  string `json:"userId"`
	Reason  string `json:"reason"`
	Amount  int    `json:"amount"`
	MatchID string `json:"matchId,omitempty"`
}

func (h *PointsHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PointsHandler) HandlePostEvent(w http.ResponseWriter, r *http.Request) {
	claims := auth.AccessClaimsFromContext(r.Context())
	if claims == nil {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req postPointsEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.UserID) == "" {
		httputil.Error(w, http.StatusBadRequest, "missing userId")
		return
	}
	isService := claims.Role == auth.RoleService && claims.Subject == "service"
	isUserSelf := !isService && (req.UserID == claims.UserID || req.UserID == claims.Subject)
	if !isService && !isUserSelf {
		httputil.Error(w, http.StatusForbidden, "userId must match authenticated user or use a service token")
		return
	}
	reason := domain.Reason(req.Reason)
	if err := h.AddTx.AddTransaction(r.Context(), usecase.AddTransactionInput{
		UserID:  req.UserID,
		Amount:  req.Amount,
		Reason:  reason,
		MatchID: req.MatchID,
	}); err != nil {
		if errors.Is(err, domain.ErrInvalidPointsRequest) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to record points")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PointsHandler) HandleGetMine(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	if uid == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	res, err := h.GetMine.GetMyPoints(r.Context(), uid)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load points")
		return
	}
	type tx struct {
		ID        string `json:"id"`
		Amount    int    `json:"amount"`
		Reason    string `json:"reason"`
		SourceKey string `json:"sourceKey,omitempty"`
		CreatedAt string `json:"createdAt"`
	}
	outTx := make([]tx, 0, len(res.Transactions))
	for _, t := range res.Transactions {
		outTx = append(outTx, tx{
			ID: t.ID, Amount: t.Amount, Reason: string(t.Reason), SourceKey: t.SourceKey,
			CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	httputil.JSON(w, http.StatusOK, map[string]any{
		"totalPoints":  res.TotalPoints,
		"transactions": outTx,
	})
}

func parseLimit(r *http.Request, def, max int) int {
	q := r.URL.Query().Get("limit")
	if q == "" {
		return def
	}
	n, err := strconv.Atoi(q)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func (h *PointsHandler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	if uid == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit := parseLimit(r, 20, 500)
	rows, err := h.GetBoard.GetLeaderboard(r.Context(), limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load leaderboard")
		return
	}
	httputil.JSON(w, http.StatusOK, rows)
}

func (h *PointsHandler) HandleGetReputationLeaderboard(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	if uid == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit := parseLimit(r, 20, 500)
	rows, err := h.GetRepBoard.GetReputationLeaderboard(r.Context(), limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load leaderboard")
		return
	}
	httputil.JSON(w, http.StatusOK, rows)
}

func (h *PointsHandler) HandleSearchLeaderboard(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	if uid == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := parseLimit(r, 10, 100)
	rows, err := h.SearchBoard.SearchLeaderboard(r.Context(), q, limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to search leaderboard")
		return
	}
	httputil.JSON(w, http.StatusOK, rows)
}

func (h *PointsHandler) HandleGetPublicUserPoints(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		httputil.Error(w, http.StatusBadRequest, "missing user id")
		return
	}
	total, err := h.GetPublicTotal.GetUserPointsTotal(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to load points")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"userId": userID, "totalPoints": total})
}
