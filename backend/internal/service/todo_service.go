package service

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/observability"
)

type todoService struct {
	todoRepo        domain.TodoRepository
	connectionRepo  domain.ConnectionRepository
	userRepo        domain.UserRepository
	gamificationSvc domain.GamificationService
	notificationSvc domain.NotificationService
	validate        *validator.Validate
}

func NewTodoService(
	todoRepo domain.TodoRepository,
	connectionRepo domain.ConnectionRepository,
	userRepo domain.UserRepository,
	gamificationSvc domain.GamificationService,
	notificationSvc domain.NotificationService,
) domain.TodoService {
	return &todoService{
		todoRepo:        todoRepo,
		connectionRepo:  connectionRepo,
		userRepo:        userRepo,
		gamificationSvc: gamificationSvc,
		notificationSvc: notificationSvc,
		validate:        validator.New(),
	}
}

func (s *todoService) Create(userID uuid.UUID, input domain.CreateTodoInput) (*domain.Todo, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, domain.ErrValidation
	}

	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)

	if len(title) < 1 || len(title) > 200 {
		return nil, domain.ErrInvalidTodoTitle
	}

	if len(description) > 2000 {
		return nil, domain.ErrInvalidTodoDescription
	}

	priority := input.Priority
	if priority == "" {
		priority = domain.TodoPriorityMedium
	}
	if !domain.ValidatePriority(priority) {
		return nil, domain.ErrValidation
	}

	if input.AssignedTo != nil {
		if err := s.validateAssignee(userID, *input.AssignedTo); err != nil {
			return nil, err
		}
	}

	todo := &domain.Todo{
		Title:       title,
		Description: description,
		Status:      domain.TodoStatusPending,
		Priority:    priority,
		CreatedBy:   userID,
		AssignedTo:  input.AssignedTo,
		DueDate:     input.DueDate,
	}

	if err := s.todoRepo.Create(todo); err != nil {
		return nil, err
	}

	if todo.AssignedTo != nil && *todo.AssignedTo != userID {
		s.queueTodoAssignedNotification(todo, userID)
	}

	return todo, nil
}

func (s *todoService) Get(userID uuid.UUID, todoID uuid.UUID) (*domain.Todo, error) {
	todo, err := s.todoRepo.GetByID(todoID)
	if err != nil {
		return nil, err
	}

	if !todo.IsAccessibleBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	return todo, nil
}

func (s *todoService) List(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	filters.UserID = &userID
	return s.todoRepo.List(filters)
}

func (s *todoService) Update(userID uuid.UUID, todoID uuid.UUID, input domain.UpdateTodoInput, version int) (*domain.Todo, error) {
	todo, err := s.todoRepo.GetByID(todoID)
	if err != nil {
		return nil, err
	}

	if !todo.IsOwnedBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	if err := s.validate.Struct(input); err != nil {
		return nil, domain.ErrValidation
	}

	if input.Title != "" {
		title := strings.TrimSpace(input.Title)
		if len(title) < 1 || len(title) > 200 {
			return nil, domain.ErrInvalidTodoTitle
		}
		todo.Title = title
	}

	if input.Description != "" {
		description := strings.TrimSpace(input.Description)
		if len(description) > 2000 {
			return nil, domain.ErrInvalidTodoDescription
		}
		todo.Description = description
	}

	if input.Status != "" {
		if !todo.CanTransitionTo(input.Status) {
			return nil, domain.ErrInvalidStatusTransition
		}
		todo.Status = input.Status
	}

	if input.Priority != "" {
		if !domain.ValidatePriority(input.Priority) {
			return nil, domain.ErrValidation
		}
		todo.Priority = input.Priority
	}

	if input.AssignedTo != nil {
		if err := s.validateAssignee(userID, *input.AssignedTo); err != nil {
			return nil, err
		}
		todo.AssignedTo = input.AssignedTo
	}

	if input.DueDate != nil {
		todo.DueDate = input.DueDate
	}

	todo.Version = version

	if err := s.todoRepo.Update(todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *todoService) Delete(userID uuid.UUID, todoID uuid.UUID) error {
	todo, err := s.todoRepo.GetByID(todoID)
	if err != nil {
		return err
	}

	if !todo.IsOwnedBy(userID) {
		return domain.ErrUnauthorized
	}

	return s.todoRepo.Delete(todoID)
}

func (s *todoService) Assign(todoID uuid.UUID, userID uuid.UUID, assignToID uuid.UUID) (*domain.Todo, error) {
	todo, err := s.todoRepo.GetByID(todoID)
	if err != nil {
		return nil, err
	}

	if !todo.IsOwnedBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	if err := s.validateAssignee(userID, assignToID); err != nil {
		return nil, err
	}

	todo.AssignedTo = &assignToID

	if err := s.todoRepo.Update(todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *todoService) Complete(userID uuid.UUID, todoID uuid.UUID, version int) (*domain.Todo, error) {
	todo, err := s.todoRepo.GetByID(todoID)
	if err != nil {
		return nil, err
	}

	if !todo.IsAccessibleBy(userID) {
		return nil, domain.ErrUnauthorized
	}

	if !todo.CanTransitionTo(domain.TodoStatusCompleted) {
		return nil, domain.ErrInvalidStatusTransition
	}

	todo.Status = domain.TodoStatusCompleted
	todo.Version = version

	if err := s.todoRepo.Update(todo); err != nil {
		return nil, err
	}

	s.triggerGamification(todo, userID)
	s.queueTodoCompletedNotification(todo, userID)
	observability.RecordTodoCompleted(userID.String())

	return todo, nil
}

func (s *todoService) validateAssignee(currentUserID, assigneeID uuid.UUID) error {
	if currentUserID == assigneeID {
		return nil
	}

	_, err := s.userRepo.GetByID(assigneeID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return domain.ErrInvalidAssignee
		}
		return err
	}

	_, err = s.connectionRepo.GetByUserPair(currentUserID, assigneeID)
	if err != nil {
		if err == domain.ErrConnectionNotFound {
			return domain.ErrUnauthorizedAction
		}
		return err
	}

	return nil
}

func (s *todoService) triggerGamification(todo *domain.Todo, userID uuid.UUID) {
	if s.gamificationSvc != nil {
		s.gamificationSvc.OnTodoCompleted(userID, todo.UpdatedAt)
	}
}

func (s *todoService) queueTodoAssignedNotification(todo *domain.Todo, assignerID uuid.UUID) {
	if s.notificationSvc == nil || todo.AssignedTo == nil {
		return
	}

	assigner, err := s.userRepo.GetByID(assignerID)
	if err != nil {
		return
	}

	s.notificationSvc.QueueTodoAssigned(
		*todo.AssignedTo,
		todo.ID,
		todo.Title,
		assignerID,
		assigner.DisplayName,
	)
}

func (s *todoService) queueTodoCompletedNotification(todo *domain.Todo, completedByID uuid.UUID) {
	if s.notificationSvc == nil {
		return
	}

	completedBy, err := s.userRepo.GetByID(completedByID)
	if err != nil {
		return
	}

	if todo.CreatedBy != completedByID {
		s.notificationSvc.QueueTodoCompleted(
			todo.CreatedBy,
			todo.ID,
			todo.Title,
			completedByID,
			completedBy.DisplayName,
		)
	}

	if todo.AssignedTo != nil && *todo.AssignedTo != completedByID && *todo.AssignedTo != todo.CreatedBy {
		s.notificationSvc.QueueTodoCompleted(
			*todo.AssignedTo,
			todo.ID,
			todo.Title,
			completedByID,
			completedBy.DisplayName,
		)
	}
}
