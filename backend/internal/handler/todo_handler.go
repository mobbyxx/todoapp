package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/middleware"
)

type TodoHandler struct {
	todoService domain.TodoService
	validate    *validator.Validate
}

func NewTodoHandler(todoService domain.TodoService) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
		validate:    validator.New(),
	}
}

type createTodoRequest struct {
	Title       string              `json:"title" validate:"required,min=1,max=200"`
	Description string              `json:"description,omitempty" validate:"omitempty,max=2000"`
	Priority    domain.TodoPriority `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
	AssignedTo  *uuid.UUID          `json:"assigned_to,omitempty"`
	DueDate     *string             `json:"due_date,omitempty"`
}

type updateTodoRequest struct {
	Title       string              `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description string              `json:"description,omitempty" validate:"omitempty,max=2000"`
	Status      domain.TodoStatus   `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    domain.TodoPriority `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
	AssignedTo  *uuid.UUID          `json:"assigned_to,omitempty"`
	DueDate     *string             `json:"due_date,omitempty"`
	Version     int                 `json:"version" validate:"required,min=1"`
}

type completeTodoRequest struct {
	Version int `json:"version" validate:"required,min=1"`
}

type todoResponse struct {
	Todo domain.TodoResponse `json:"todo"`
}

type todoListResponse struct {
	Todos      []domain.TodoResponse `json:"todos"`
	TotalCount int                   `json:"total_count"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	input := domain.CreateTodoInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		AssignedTo:  req.AssignedTo,
	}

	if req.DueDate != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Due date parsing not implemented in this version")
		return
	}

	todo, err := h.todoService.Create(userID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid input data")
		case errors.Is(err, domain.ErrInvalidTodoTitle):
			writeError(w, http.StatusBadRequest, "validation_error", "Title must be 1-200 characters")
		case errors.Is(err, domain.ErrInvalidTodoDescription):
			writeError(w, http.StatusBadRequest, "validation_error", "Description must be max 2000 characters")
		case errors.Is(err, domain.ErrInvalidAssignee):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid assignee")
		case errors.Is(err, domain.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "forbidden", "Can only assign to connections")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create todo")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todoResponse{
		Todo: todo.ToResponse(),
	})
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
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

	filters := domain.TodoFilters{
		Page:     1,
		PageSize: 20,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.TodoStatus(status)
		if domain.ValidateStatus(s) {
			filters.Status = &s
		}
	}

	if priority := r.URL.Query().Get("priority"); priority != "" {
		p := domain.TodoPriority(priority)
		if domain.ValidatePriority(p) {
			filters.Priority = &p
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			filters.PageSize = ps
		}
	}

	todos, totalCount, err := h.todoService.List(userID, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list todos")
		return
	}

	response := todoListResponse{
		Todos:      make([]domain.TodoResponse, len(todos)),
		TotalCount: totalCount,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
	}

	for i, todo := range todos {
		response.Todos[i] = todo.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TodoHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	todoIDStr := chi.URLParam(r, "id")
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid todo ID")
		return
	}

	todo, err := h.todoService.Get(userID, todoID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTodoNotFound):
			writeError(w, http.StatusNotFound, "todo_not_found", "Todo not found")
		case errors.Is(err, domain.ErrUnauthorized):
			writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get todo")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoResponse{
		Todo: todo.ToResponse(),
	})
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	todoIDStr := chi.URLParam(r, "id")
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid todo ID")
		return
	}

	var req updateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	input := domain.UpdateTodoInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssignedTo:  req.AssignedTo,
	}

	if req.DueDate != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Due date parsing not implemented in this version")
		return
	}

	todo, err := h.todoService.Update(userID, todoID, input, req.Version)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTodoNotFound):
			writeError(w, http.StatusNotFound, "todo_not_found", "Todo not found")
		case errors.Is(err, domain.ErrUnauthorized):
			writeError(w, http.StatusForbidden, "forbidden", "Only the owner can update this todo")
		case errors.Is(err, domain.ErrValidation):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid input data")
		case errors.Is(err, domain.ErrInvalidTodoTitle):
			writeError(w, http.StatusBadRequest, "validation_error", "Title must be 1-200 characters")
		case errors.Is(err, domain.ErrInvalidTodoDescription):
			writeError(w, http.StatusBadRequest, "validation_error", "Description must be max 2000 characters")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid status transition")
		case errors.Is(err, domain.ErrVersionMismatch):
			writeError(w, http.StatusConflict, "version_mismatch", "Todo was modified by another request. Please refresh and try again.")
		case errors.Is(err, domain.ErrInvalidAssignee):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid assignee")
		case errors.Is(err, domain.ErrUnauthorizedAction):
			writeError(w, http.StatusForbidden, "forbidden", "Can only assign to connections")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update todo")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoResponse{
		Todo: todo.ToResponse(),
	})
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	todoIDStr := chi.URLParam(r, "id")
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid todo ID")
		return
	}

	if err := h.todoService.Delete(userID, todoID); err != nil {
		switch {
		case errors.Is(err, domain.ErrTodoNotFound):
			writeError(w, http.StatusNotFound, "todo_not_found", "Todo not found")
		case errors.Is(err, domain.ErrUnauthorized):
			writeError(w, http.StatusForbidden, "forbidden", "Only the owner can delete this todo")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete todo")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Todo deleted successfully",
	})
}

func (h *TodoHandler) Complete(w http.ResponseWriter, r *http.Request) {
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

	todoIDStr := chi.URLParam(r, "id")
	todoID, err := uuid.Parse(todoIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid todo ID")
		return
	}

	var req completeTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	todo, err := h.todoService.Complete(userID, todoID, req.Version)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTodoNotFound):
			writeError(w, http.StatusNotFound, "todo_not_found", "Todo not found")
		case errors.Is(err, domain.ErrUnauthorized):
			writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid status transition")
		case errors.Is(err, domain.ErrVersionMismatch):
			writeError(w, http.StatusConflict, "version_mismatch", "Todo was modified by another request. Please refresh and try again.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to complete todo")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoResponse{
		Todo: todo.ToResponse(),
	})
}

func (h *TodoHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/complete", h.Complete)

	return r
}
