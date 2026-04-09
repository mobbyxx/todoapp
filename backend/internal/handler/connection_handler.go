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

type ConnectionHandler struct {
	service domain.ConnectionService
}

func NewConnectionHandler(service domain.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{
		service: service,
	}
}

func (h *ConnectionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/connections", func(r chi.Router) {
		r.Post("/invite", h.CreateInvitation)
		r.Get("/invite/{token}", h.ValidateInvitation)
		r.Post("/invite/{token}/accept", h.AcceptInvitation)
		r.Post("/invite/{token}/reject", h.RejectInvitation)
		r.Get("/", h.ListConnections)
		r.Delete("/{connectionID}", h.RemoveConnection)

		r.Post("/qrcode/generate", h.GenerateQRCode)
		r.Post("/qrcode/scan", h.ScanQRCode)
	})
}

func (h *ConnectionHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	connection, token, err := h.service.CreateInvitation(userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectionAlreadyExists):
			writeError(w, http.StatusConflict, "conflict", "connection already exists")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to create invitation")
		}
		return
	}

	response := domain.InvitationResponse{
		ConnectionID:   connection.ID.String(),
		Token:          token,
		InvitationLink: domain.GenerateInvitationLink(token),
		ExpiresAt:      *connection.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ConnectionHandler) ValidateInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "token is required")
		return
	}

	connection, err := h.service.ValidateInvitation(token)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInvitationToken):
			writeError(w, http.StatusNotFound, "invalid_token", "invalid invitation token")
		case errors.Is(err, domain.ErrInvitationExpired):
			writeError(w, http.StatusGone, "expired", "invitation has expired")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to validate invitation")
		}
		return
	}

	response := map[string]interface{}{
		"connection_id": connection.ID.String(),
		"requested_by":  connection.RequestedBy.String(),
		"status":        connection.Status,
		"expires_at":    connection.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ConnectionHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "token is required")
		return
	}

	if err := h.service.AcceptInvitation(userID, token); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInvitationToken):
			writeError(w, http.StatusNotFound, "invalid_token", "invalid invitation token")
		case errors.Is(err, domain.ErrInvitationExpired):
			writeError(w, http.StatusGone, "expired", "invitation has expired")
		case errors.Is(err, domain.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "forbidden", "cannot accept your own invitation")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			writeError(w, http.StatusConflict, "conflict", "invitation cannot be accepted")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to accept invitation")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConnectionHandler) RejectInvitation(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "token is required")
		return
	}

	if err := h.service.RejectInvitation(userID, token); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInvitationToken):
			writeError(w, http.StatusNotFound, "invalid_token", "invalid invitation token")
		case errors.Is(err, domain.ErrInvitationExpired):
			writeError(w, http.StatusGone, "expired", "invitation has expired")
		case errors.Is(err, domain.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "forbidden", "cannot reject your own invitation")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			writeError(w, http.StatusConflict, "conflict", "invitation cannot be rejected")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to reject invitation")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConnectionHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	connections, err := h.service.GetConnections(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list connections")
		return
	}

	response := make([]domain.ExtendedConnectionResponse, len(connections))
	for i, conn := range connections {
		response[i] = domain.ExtendedConnectionResponse{
			ID:          conn.ID.String(),
			UserAID:     conn.UserAID.String(),
			UserBID:     conn.UserBID.String(),
			Status:      string(conn.Status),
			RequestedBy: conn.RequestedBy.String(),
			AcceptedAt:  conn.AcceptedAt,
			CreatedAt:   conn.CreatedAt,
			UpdatedAt:   conn.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ConnectionHandler) RemoveConnection(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid user ID")
		return
	}

	connectionIDStr := chi.URLParam(r, "connectionID")
	connectionID, err := uuid.Parse(connectionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid connection ID")
		return
	}

	if err := h.service.Disconnect(connectionID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrConnectionNotFound):
			writeError(w, http.StatusNotFound, "not_found", "connection not found")
		case errors.Is(err, domain.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "forbidden", "not authorized to remove this connection")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to remove connection")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConnectionHandler) GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "deprecated", "QR code generation is deprecated, use /api/v1/connections/invite instead")
}

func (h *ConnectionHandler) ScanQRCode(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "deprecated", "QR code scanning is deprecated, use invitation links instead")
}
