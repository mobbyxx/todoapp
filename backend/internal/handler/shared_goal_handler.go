package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/middleware"
)

type SharedGoalHandler struct {
	sharedGoalService domain.SharedGoalService
	connectionService domain.ConnectionService
}

func NewSharedGoalHandler(sharedGoalService domain.SharedGoalService, connectionService domain.ConnectionService) *SharedGoalHandler {
	return &SharedGoalHandler{
		sharedGoalService: sharedGoalService,
		connectionService: connectionService,
	}
}

func (h *SharedGoalHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
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

	var req domain.CreateSharedGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	connectionID, err := uuid.Parse(req.ConnectionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_connection_id", "Invalid connection ID")
		return
	}

	if !h.isUserInConnection(userID, connectionID) {
		writeError(w, http.StatusForbidden, "forbidden", "User is not a participant in this connection")
		return
	}

	targetType := domain.SharedGoalTargetType(req.TargetType)
	if err := domain.ValidateTargetType(targetType); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_target_type", "Invalid target type")
		return
	}

	if err := domain.ValidateTargetValue(req.TargetValue); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_target_value", "Target value must be greater than 0")
		return
	}

	goal, err := h.sharedGoalService.CreateGoal(connectionID, targetType, req.TargetValue, req.RewardDescription)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidTargetType):
			writeError(w, http.StatusBadRequest, "invalid_target_type", err.Error())
		case errors.Is(err, domain.ErrInvalidTargetValue):
			writeError(w, http.StatusBadRequest, "invalid_target_value", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create shared goal")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(goal.ToResponse())
}

func (h *SharedGoalHandler) ListGoals(w http.ResponseWriter, r *http.Request) {
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

	goals, err := h.sharedGoalService.ListGoals(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list shared goals")
		return
	}

	response := domain.ListSharedGoalsResponse{
		Goals: make([]domain.SharedGoalResponse, len(goals)),
	}
	for i, goal := range goals {
		response.Goals[i] = goal.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SharedGoalHandler) isUserInConnection(userID, connectionID uuid.UUID) bool {
	connections, err := h.connectionService.GetConnections(userID)
	if err != nil {
		return false
	}

	for _, conn := range connections {
		if conn.ID == connectionID {
			return true
		}
	}

	return false
}

func (h *SharedGoalHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateGoal)
	r.Get("/", h.ListGoals)

	return r
}
