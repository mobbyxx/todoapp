package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/middleware"
)

type GamificationHandler struct {
	gamificationService domain.GamificationService
}

func NewGamificationHandler(gamificationService domain.GamificationService) *GamificationHandler {
	return &GamificationHandler{
		gamificationService: gamificationService,
	}
}

type userStatsResponse struct {
	Points              int     `json:"points"`
	Level               int     `json:"level"`
	LevelName           string  `json:"level_name"`
	Streak              int     `json:"streak"`
	TotalTodosCompleted int     `json:"total_todos_completed"`
	ProgressToNextLevel float64 `json:"progress_to_next_level"`
	XPForNextLevel      int     `json:"xp_for_next_level"`
}

type pointsHistoryResponse struct {
	History []*domain.PointsTransaction `json:"history"`
}

func (h *GamificationHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid user ID")
		return
	}

	stats, err := h.gamificationService.GetUserStats(userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get user stats")
		}
		return
	}

	response := userStatsResponse{
		Points:              stats.Points,
		Level:               stats.Level,
		LevelName:           domain.GetLevelName(stats.Level),
		Streak:              stats.Streak,
		TotalTodosCompleted: stats.TotalTodosCompleted,
		ProgressToNextLevel: domain.GetLevelProgress(stats.Points),
		XPForNextLevel:      domain.GetXPForNextLevel(stats.Points),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *GamificationHandler) GetPointsHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid user ID")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	history, err := h.gamificationService.GetPointsHistory(userID, limit)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get points history")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pointsHistoryResponse{
		History: history,
	})
}

func (h *GamificationHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/users/me/stats", h.GetUserStats)
	r.Get("/users/me/history", h.GetPointsHistory)

	return r
}
