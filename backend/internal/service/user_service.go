package service

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/security"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

type userService struct {
	repo           domain.UserRepository
	jwtService     *JWTService
	validate       *validator.Validate
	passwordPolicy *security.PasswordPolicy
}

func NewUserService(repo domain.UserRepository, jwtService *JWTService) domain.UserService {
	return &userService{
		repo:           repo,
		jwtService:     jwtService,
		validate:       validator.New(),
		passwordPolicy: security.DefaultPasswordPolicy(),
	}
}

func (s *userService) Register(email, password, displayName string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	displayName = strings.TrimSpace(displayName)

	req := domain.RegistrationRequest{
		Email:       email,
		Password:    password,
		DisplayName: displayName,
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, domain.ErrValidation
	}

	if err := s.passwordPolicy.Validate(password, email); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetByEmail(email); err == nil {
		return nil, domain.ErrEmailTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  displayName,
		IsActive:     true,
		Preferences:  make(map[string]interface{}),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	req := domain.LoginRequest{
		Email:    email,
		Password: password,
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, domain.ErrValidation
	}

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	_ = s.repo.UpdateLastSeen(user.ID)

	return user, nil
}

func (s *userService) GetUser(id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetUserByEmail(email string) (*domain.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *userService) UpdateProfile(id uuid.UUID, displayName string) (*domain.User, error) {
	displayName = strings.TrimSpace(displayName)

	if len(displayName) < 2 || len(displayName) > 50 {
		return nil, domain.ErrInvalidDisplayName
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	user.DisplayName = displayName

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateLastSeen(id uuid.UUID) error {
	return s.repo.UpdateLastSeen(id)
}

func (s *userService) SoftDelete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
