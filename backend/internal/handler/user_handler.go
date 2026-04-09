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
	"github.com/user/todo-api/internal/service"
)

type UserHandler struct {
	userService domain.UserService
	jwtService  *service.JWTService
	validate    *validator.Validate
}

func NewUserHandler(userService domain.UserService, jwtService *service.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
		validate:    validator.New(),
	}
}

type registerRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name" validate:"omitempty,min=2,max=50"`
}

type authResponse struct {
	User         domain.UserProfile `json:"user"`
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	ExpiresAt    string             `json:"expires_at"`
}

type userResponse struct {
	User domain.UserProfile `json:"user"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := h.userService.Register(req.Email, req.Password, req.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email_taken", "Email already registered")
		case errors.Is(err, domain.ErrValidation):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid input data")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create user")
		}
		return
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(r.Context(), user.ID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate tokens")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse{
		User:         user.ToProfile(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to login")
		}
		return
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(r.Context(), user.ID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to generate tokens")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{
		User:         user.ToProfile(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	tokenPair, err := h.jwtService.ValidateRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenExpired):
			writeError(w, http.StatusUnauthorized, "token_expired", "Refresh token has expired")
		case errors.Is(err, service.ErrTokenRevoked):
			writeError(w, http.StatusUnauthorized, "token_revoked", "Refresh token has been revoked")
		case errors.Is(err, service.ErrTokenVersionMismatch):
			writeError(w, http.StatusUnauthorized, "token_invalid", "Invalid refresh token")
		default:
			writeError(w, http.StatusUnauthorized, "invalid_token", "Invalid refresh token")
		}
		return
	}

	userID, _ := uuid.Parse(req.RefreshToken)
	_ = userID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.jwtService.RevokeToken(r.Context(), req.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully logged out",
	})
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.userService.GetUser(userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get user")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponse{
		User: user.ToProfile(),
	})
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
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

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Display name is required")
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user_not_found", "User not found")
		case errors.Is(err, domain.ErrInvalidDisplayName):
			writeError(w, http.StatusBadRequest, "validation_error", "Display name must be 2-50 characters")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update profile")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponse{
		User: user.ToProfile(),
	})
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)

	return r
}

func (h *UserHandler) ProtectedRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/me", h.GetMe)
	r.Put("/me", h.UpdateMe)

	return r
}
