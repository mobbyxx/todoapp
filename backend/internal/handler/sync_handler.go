package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/middleware"
)

type SyncHandler struct {
	syncService     domain.SyncService
	conflictService domain.ConflictService
	validate        *validator.Validate
}

func NewSyncHandler(syncService domain.SyncService, conflictService domain.ConflictService) *SyncHandler {
	return &SyncHandler{
		syncService:     syncService,
		conflictService: conflictService,
		validate:        validator.New(),
	}
}

type syncRequest struct {
	LastPulledAt domain.HLC       `json:"last_pulled_at"`
	Changes      domain.ChangeSet `json:"changes,omitempty"`
}

type syncResponse struct {
	Timestamp  domain.HLC             `json:"timestamp"`
	Changes    domain.ChangeSet       `json:"changes"`
	Conflicts  []*domain.SyncConflict `json:"conflicts"`
	Status     domain.SyncStatus      `json:"status"`
	ServerTime int64                  `json:"server_time"`
}

func (h *SyncHandler) Sync(w http.ResponseWriter, r *http.Request) {
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

	var req syncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	syncReq := domain.SyncRequest{
		LastPulledAt: req.LastPulledAt,
		Changes:      req.Changes,
	}

	result, err := h.syncService.Sync(userID, syncReq)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidChangeSet):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid change set")
		case errors.Is(err, domain.ErrSyncFailed):
			writeError(w, http.StatusInternalServerError, "sync_failed", "Sync operation failed")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to sync")
		}
		return
	}

	response := syncResponse{
		Timestamp:  result.Timestamp,
		Changes:    result.Changes,
		Conflicts:  result.Conflicts,
		Status:     result.Status,
		ServerTime: result.ServerTime,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *SyncHandler) GetLastSync(w http.ResponseWriter, r *http.Request) {
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

	lastSync, err := h.syncService.GetLastSync(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get last sync")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"last_synced_at": lastSync,
	})
}

func (h *SyncHandler) GetUnresolvedConflicts(w http.ResponseWriter, r *http.Request) {
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

	conflicts, err := h.conflictService.GetUnresolvedConflicts(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get conflicts")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conflicts": conflicts,
	})
}

func (h *SyncHandler) ResolveConflict(w http.ResponseWriter, r *http.Request) {
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

	conflictIDStr := chi.URLParam(r, "conflictID")
	conflictID, err := uuid.Parse(conflictIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid conflict ID")
		return
	}

	var req domain.ConflictResolution
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	req.ConflictID = conflictID

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.conflictService.ResolveConflictManually(conflictID, req, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "forbidden", "Not authorized to resolve this conflict")
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "Conflict not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to resolve conflict")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Conflict resolved successfully",
	})
}

func (h *SyncHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Sync)
	r.Get("/last", h.GetLastSync)
	r.Get("/conflicts", h.GetUnresolvedConflicts)
	r.Post("/conflicts/{conflictID}/resolve", h.ResolveConflict)

	return r
}
