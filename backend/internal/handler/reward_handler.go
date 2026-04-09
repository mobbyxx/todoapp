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

type RewardHandler struct {
	rewardService domain.RewardService
	validate      *validator.Validate
}

func NewRewardHandler(rewardService domain.RewardService) *RewardHandler {
	return &RewardHandler{
		rewardService: rewardService,
		validate:      validator.New(),
	}
}

type createRewardRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
	Cost        int    `json:"cost" validate:"required,min=1"`
}

type rewardResponse struct {
	Reward domain.Reward `json:"reward"`
}

type rewardsListResponse struct {
	Rewards []domain.Reward `json:"rewards"`
}

type redemptionResponse struct {
	Redemption domain.RewardRedemption `json:"redemption"`
}

type redemptionsListResponse struct {
	Redemptions []domain.RewardRedemption `json:"redemptions"`
}

func (h *RewardHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	input := domain.CreateRewardInput{
		Name:        req.Name,
		Description: req.Description,
		Cost:        req.Cost,
	}

	reward, err := h.rewardService.CreateReward(userID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRewardName):
			writeError(w, http.StatusBadRequest, "validation_error", "Reward name must be 1-100 characters")
		case errors.Is(err, domain.ErrInvalidRewardDescription):
			writeError(w, http.StatusBadRequest, "validation_error", "Reward description must be max 500 characters")
		case errors.Is(err, domain.ErrInvalidRewardCost):
			writeError(w, http.StatusBadRequest, "validation_error", "Reward cost must be positive")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create reward")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rewardResponse{
		Reward: *reward,
	})
}

func (h *RewardHandler) List(w http.ResponseWriter, r *http.Request) {
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

	rewards, err := h.rewardService.GetRewards(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list rewards")
		return
	}

	response := rewardsListResponse{
		Rewards: make([]domain.Reward, len(rewards)),
	}
	for i, reward := range rewards {
		response.Rewards[i] = *reward
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RewardHandler) Redeem(w http.ResponseWriter, r *http.Request) {
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

	rewardIDStr := chi.URLParam(r, "id")
	rewardID, err := uuid.Parse(rewardIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid reward ID")
		return
	}

	redemption, err := h.rewardService.RedeemReward(userID, rewardID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRewardNotFound):
			writeError(w, http.StatusNotFound, "reward_not_found", "Reward not found")
		case errors.Is(err, domain.ErrRewardNotActive):
			writeError(w, http.StatusBadRequest, "reward_inactive", "Reward is not active")
		case errors.Is(err, domain.ErrInsufficientXP):
			writeError(w, http.StatusBadRequest, "insufficient_xp", "Not enough XP to redeem this reward")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to redeem reward")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redemptionResponse{
		Redemption: *redemption,
	})
}

func (h *RewardHandler) GetMyRedemptions(w http.ResponseWriter, r *http.Request) {
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

	redemptions, err := h.rewardService.GetMyRewards(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get redemptions")
		return
	}

	response := redemptionsListResponse{
		Redemptions: make([]domain.RewardRedemption, len(redemptions)),
	}
	for i, redemption := range redemptions {
		response.Redemptions[i] = *redemption
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RewardHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/rewards", h.Create)
	r.Get("/rewards", h.List)
	r.Post("/rewards/{id}/redeem", h.Redeem)
	r.Get("/rewards/my", h.GetMyRedemptions)

	return r
}
